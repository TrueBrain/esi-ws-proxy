package main

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebServer struct {
}

func NewWebSocket() *WebServer {
	return &WebServer{}
}

func (ws *WebServer) Run() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Failed to upgrade to WebSocket", "error", err)
			return
		}

		client := newWebSocketClient(conn)
		go client.run()
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client/index.html")
	})
	http.HandleFunc("/demo.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "client/demo.html")
	})

	http.ListenAndServe(":8080", nil)
}
