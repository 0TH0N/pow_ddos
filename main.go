package main

import (
	"context"
	"fmt"
	"powserver/client"
	"powserver/server"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	clientRunner := client.NewRunner()
	go clientRunner.Run(ctx)

	srv := server.NewServer()
	err := srv.Start()
	if err != nil {
		fmt.Println("SERVER:", err)
	}

	cancel()
}
