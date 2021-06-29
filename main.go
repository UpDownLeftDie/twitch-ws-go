package main

import (
	"os"

	"github.com/updownleftdie/twitch-ws-go/v2/cmd"
)

func main() {
	exit := make(chan string)
	cmd.Execute()
	select {
	case <-exit:
		os.Exit(0)
	}
}
