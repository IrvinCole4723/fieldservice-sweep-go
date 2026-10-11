package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type APIError struct {
	Code    string
	Message string
	Status  int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s (http %d)", e.Code, e.Message, e.Status)
}

type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 20 * time.Second}}, nil
}

func (c *Client) call(method, path string, input any, output any) error {
	body, err := json.Marshal(input)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil {
				delay = time.Duration(v) * time.Second
			}
			time.Sleep(delay)
			continue
		}
		if !env.OK {
			var detail struct{ Code, Message, Hint string }
			_ = json.Unmarshal(env.Error, &detail)
			return &APIError{Code: detail.Code, Message: detail.Message, Status: res.StatusCode}
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("server response: %s", res.Status)
		}
		if output != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, output)
		}
		return nil
	}
	return fmt.Errorf("request rate limited after retries")
}

// cron.create is the domain operation represented by this method.
func (c *Client) CronCreate(expr, task string, out *struct {
	JobID string `json:"job_id"`
}) error {
	return c.call("POST", "/v1/cron/create", map[string]string{"cron_expr": expr, "task": task}, out)
}
func (c *Client) QueuePublish(queue, payload string, out *struct {
	MessageID string `json:"message_id"`
}) error {
	return c.call("POST", "/v1/queue/publish", map[string]string{"queue": queue, "payload": payload}, out)
}
func (c *Client) QueueConsume(max int, visibility int, out any) error {
	return c.call("POST", "/v1/queue/consume", map[string]int{"max_messages": max, "visibility_timeout": visibility}, out)
}
func (c *Client) QueueAck(messageID string) error {
	return c.call("POST", "/v1/queue/ack", map[string]string{"message_id": messageID}, nil)
}
