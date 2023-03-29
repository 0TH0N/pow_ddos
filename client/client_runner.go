package client

import (
	"context"
	"fmt"
	"time"
)

const (
	Scheme = "http"
	Url    = "localhost"
	Port   = 8081
	Path   = "random_quote"
)

type PowClient interface {
	GetTask() (string, error)
	Compute(string) (string, error)
	SendSolution(string, string) (string, error)
}

func NewRunner() *Runner {
	return &Runner{
		client: NewClient(Scheme, Url, Port, Path),
	}
}

type Runner struct {
	client PowClient
}

func (r *Runner) Run(ctx context.Context) {
	for true {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(time.Second * 2):
			task, err := r.client.GetTask()
			if err != nil {
				fmt.Println("CLIENT:", err)
				fmt.Println()
				continue
			}
			fmt.Println("CLIENT:", "got task -", task)

			t := time.Now()
			hash, err := r.client.Compute(task)
			if err != nil {
				fmt.Println("CLIENT:", err)
				fmt.Println()
				continue
			}
			fmt.Println("CLIENT:", "computed hash -", hash)
			fmt.Println("CLIENT:", "computed for - ", time.Since(t))

			t = time.Now()
			resp, err := r.client.SendSolution(task, hash)
			if err != nil {
				fmt.Println("CLIENT:", err)
				fmt.Println()
				continue
			}
			fmt.Println("CLIENT:", "wisdom quote -", resp)
			fmt.Println("CLIENT:", "server answered for - ", time.Since(t))
			fmt.Println()
		}
	}
}
