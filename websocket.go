package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/eveshipfit/esi-ws-proxy/queue"
)

type esiMessage struct {
	callback func(sendBuffer chan *serverMessage, token string, userData any) *time.Time
	schedule time.Time
	userData any
}

type webSocketClient struct {
	sendBuffer chan *serverMessage
	esiChan    chan *esiMessage

	conn          *websocket.Conn
	token         *string
	characterName string
	characterId   string

	esiQueue queue.PriorityQueue
}

type clientMessageType interface {
	handle(c *webSocketClient)
}

func (m *clientAuthenticate) handle(c *webSocketClient) {
	token, err := jwt.ParseString(m.Token, jwt.WithVerify(false))

	if err != nil {
		c.sendBuffer <- &serverMessage{Error: &serverError{Error: "Invalid token"}}
		return
	}

	if token.Issuer() != "https://login.eveonline.com" {
		c.sendBuffer <- &serverMessage{Error: &serverError{Error: "Token is not an EVE Online token"}}
		return
	}

	c.token = &m.Token
	if name, ok := token.Get("name"); ok {
		c.characterName = name.(string)
	}
	/* The subject is like "CHARACTER:EVE:1234567890". The CharacterID is the last part. */
	c.characterId = strings.Split(token.Subject(), ":")[2]
}
func (m *clientCharacterGetId) handle(c *webSocketClient) {
	if c.token == nil {
		c.sendBuffer <- &serverMessage{Error: &serverError{Error: "Not authenticated"}}
		return
	}

	c.sendBuffer <- &serverMessage{Character: &serverCharacter{Id: c.characterId}}
}

type LocationInfo struct {
	characterId   string
	solarSystemId int
}

func (m *clientLocationSubscribe) handle(c *webSocketClient) {
	if c.token == nil {
		c.sendBuffer <- &serverMessage{Error: &serverError{Error: "Not authenticated"}}
		return
	}

	c.esiChan <- &esiMessage{
		callback: fetchLocation,
		schedule: time.Now(),
		userData: &LocationInfo{
			characterId:   c.characterId,
			solarSystemId: 0,
		},
	}
}

func fetchLocation(sendBuffer chan *serverMessage, token string, userData any) *time.Time {
	locationInfo := userData.(*LocationInfo)

	response, err := makeEsiRequest(token, "v2/characters/"+locationInfo.characterId+"/location/")
	if err != nil {
		slog.Error("ESI request failed", "error", err)
		sendBuffer <- &serverMessage{Error: &serverError{Error: "Failed to fetch location; subscription cancelled"}}
		return nil
	}

	var res struct {
		SolarSystemId int `json:"solar_system_id"`
	}
	if err := json.Unmarshal([]byte(response), &res); err != nil {
		slog.Error("Failed to parse ESI response", "error", err)
		sendBuffer <- &serverMessage{Error: &serverError{Error: "Failed to fetch location; subscription cancelled"}}
		return nil
	}

	if res.SolarSystemId != locationInfo.solarSystemId {
		sendBuffer <- &serverMessage{Location: &serverLocation{SolarSystemId: res.SolarSystemId}}
		locationInfo.solarSystemId = res.SolarSystemId
	}

	nextTime := time.Now().Add(5 * time.Second)
	return &nextTime
}

func newWebSocketClient(conn *websocket.Conn) webSocketClient {
	return webSocketClient{
		conn:       conn,
		sendBuffer: make(chan *serverMessage, 10),
		esiChan:    make(chan *esiMessage, 10),
	}
}

/**
 * Call handle() on the members of the struct marked as "esiws" that are not nil.
 *
 * If the "esiws" tag is set to "nested", it will repeat this process for the members of that struct.
 */
func processEsiWsMessage(c *webSocketClient, msg any) {
	values := reflect.ValueOf(msg)
	types := reflect.TypeOf(msg)

	if values.Kind() == reflect.Ptr {
		values = values.Elem()
		types = types.Elem()
	}

	for i := 0; i < values.NumField(); i++ {
		value := values.Field(i)
		t := types.Field(i)
		if value.IsNil() {
			continue
		}

		if esiWs, ok := t.Tag.Lookup("esiws"); ok {
			if esiWs == "nested" {
				processEsiWsMessage(c, value.Interface())
			} else {
				value.Interface().(clientMessageType).handle(c)
			}
		}
	}
}

func (c *webSocketClient) runWriter() {
	for {
		msg := <-c.sendBuffer
		if msg == nil {
			return
		}

		if err := c.conn.WriteJSON(msg); err != nil {
			slog.Error("Writing websocket failed", "error", err)
			return
		}
	}
}

func (c *webSocketClient) runEsi() {
	for {
		var wait time.Duration
		next := c.esiQueue.Peek()
		if next == nil {
			wait = time.Hour
		} else {
			wait = time.Until(next.(*esiMessage).schedule)
		}

		select {
		case <-time.After(wait):
			msg := c.esiQueue.Pop().(*esiMessage)
			nextTime := msg.callback(c.sendBuffer, *c.token, msg.userData)
			if nextTime != nil {
				msg.schedule = *nextTime
				c.esiQueue.Push(msg, int(msg.schedule.Unix()))
			}
		case msg := <-c.esiChan:
			if msg == nil {
				return
			}

			c.esiQueue.Push(msg, int(msg.schedule.Unix()))
		}
	}
}

func (c *webSocketClient) run() {
	defer c.conn.Close()
	defer func() { c.sendBuffer <- nil }()
	defer func() { c.esiChan <- nil }()
	defer func() { slog.Info("Client disconnected") }()
	slog.Info("Client connected")

	go c.runWriter()
	go c.runEsi()

	for {
		var msg clientMessage
		if err := c.conn.ReadJSON(&msg); err != nil {
			if err != io.EOF {
				slog.Error("Reading websocket failed", "error", err)
			}
			return
		}

		processEsiWsMessage(c, msg)
	}
}
