package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	addr  string
	token string
}

func New(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c *Client) SetToken(token string) {
	c.token = token
}

type Options struct {
	QueryParams map[string]string
}

func createUrl(addr, path string, query map[string]string) (string, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return "", err
	}

	u.Path = path

	params := u.Query()
	for k, v := range query {
		params.Set(k, v)
	}
	u.RawQuery = params.Encode()

	return u.String(), nil
}

type ApiError[E any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  E      `json:"errors,omitempty"`
}

func (err *ApiError[E]) Error() string {
	return err.Message
}

type ApiResponse[D any, E any] struct {
	Status string       `json:"status"`
	Data   D            `json:"data,omitempty"`
	Error  *ApiError[E] `json:"error,omitempty"`
}

type RequestData struct {
	Url    string
	Method string

	Token string
	Body  any
}

func Request[D any](data RequestData) (*D, error) {
	var bodyReader io.Reader

	if data.Body != nil {
		buf := bytes.Buffer{}

		err := json.NewEncoder(&buf).Encode(data.Body)
		if err != nil {
			return nil, err
		}

		bodyReader = &buf
	}

	req, err := http.NewRequest(data.Method, data.Url, bodyReader)
	if err != nil {
		return nil, err
	}

	if data.Token != "" {
		req.Header.Add("Authorization", "Bearer "+data.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res ApiResponse[D, any]
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	if res.Status == "error" {
		return nil, res.Error
	}
	

	return &res.Data, nil
}

// Simple wrapper for Sprintf
func Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

