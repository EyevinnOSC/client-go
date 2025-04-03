package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

type Auth struct {
	Header string
	Value  string
}

func Fetch(url string, method string, body *bytes.Buffer, auth Auth) (*http.Response, error) {
	client := &http.Client{}
	if body == nil {
		body = &bytes.Buffer{}
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	//req.Header.Set("x-pat-jwt", fmt.Sprintf("Bearer %s", token))
	req.Header.Set(auth.Header, auth.Value)
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

func FetchBytes(url string, method string, body *bytes.Buffer, auth Auth) ([]byte, error) {

	resp, err := Fetch(url, method, body, auth)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return responseBytes, nil
}

func FetchJson(url string, method string, body *bytes.Buffer, target interface{}, auth Auth) error {

	resp, err := Fetch(url, method, body, auth)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	slog.Debug(fmt.Sprintf("Response status: %d", resp.StatusCode))
	//slog.Info(fmt.Sprintf("Response body: %s", string(responseBytes)))
	if target != nil {
		value, ok := resp.Header["Content-Type"]
		if ok && strings.HasPrefix(value[0], "application/json") {
			if err := json.Unmarshal(responseBytes, target); err != nil {
				responseBody := string(responseBytes)
				slog.Warn("Unmarshall response failed", "url", url, "responsebody", responseBody)
				return err
			}
		}
	}

	return nil
}
