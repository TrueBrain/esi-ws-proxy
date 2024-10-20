package main

import (
	"fmt"
	"io"
	"net/http"
)

func makeEsiRequest(bearerToken string, endpoint string) (string, error) {
	request, _ := http.NewRequest("GET", "https://esi.evetech.net/"+endpoint, nil)
	request.Header.Set("User-Agent", "eveship.fit ESI WebSocket Proxy")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+bearerToken)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("ESI returned status code %d: %s", response.StatusCode, string(body))
	}

	return string(body), nil
}
