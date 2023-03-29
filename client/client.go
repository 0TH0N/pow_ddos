package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PoW-HC/hashcash/pkg/hash"
	"github.com/PoW-HC/hashcash/pkg/pow"
	"io"
	"net/http"
	"powserver/server"
)

func NewClient(scheme, url string, port int, path string) PowClient {
	return &Client{
		scheme: scheme,
		url:    url,
		port:   port,
		path:   path,
	}
}

type Client struct {
	scheme, url, path string
	port              int
}

func (c *Client) GetTask() (string, error) {
	requestURL := fmt.Sprintf("%s://%s:%d/%s", c.scheme, c.url, c.port, c.path)
	res, err := http.Get(requestURL)
	if err != nil {
		return "", err
	}

	task := res.Header.Get(server.TaskHeaderName)
	if len(task) == 0 {
		return "", errors.New("empty task header")
	}

	return task, nil
}

func (c *Client) Compute(task string) (string, error) {
	hasher, err := hash.NewHasher("sha256")
	if err != nil {
		return "", err
	}

	p := pow.New(hasher)

	hashCash, err := pow.InitHashcash(5, task, pow.SignExt(server.SecretPhrase, hasher))
	if err != nil {
		return "", err
	}

	solution, err := p.Compute(context.Background(), hashCash, server.MaxIterations)
	if err != nil {
		return "", err
	}

	return solution.String(), nil
}

func (c *Client) SendSolution(task, hash string) (string, error) {
	requestURL := fmt.Sprintf("%s://%s:%d/%s", c.scheme, c.url, c.port, c.path)
	data := []byte("")
	reader := bytes.NewReader(data)
	req, err := http.NewRequest("GET", requestURL, reader)
	req.Header.Set(server.TaskHeaderName, task)
	req.Header.Set(server.HashHeaderName, hash)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if len(body) == 0 {
		return "", errors.New("empty body response")
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	return string(body), nil
}
