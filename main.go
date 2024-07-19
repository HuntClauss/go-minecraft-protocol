package main

import (
	"log"
	"mc-bot/mc"

	"github.com/huntclauss/dotenv"
)

func main() {
	if err := dotenv.LoadEnv(".env"); err != nil {
		log.Fatalf("cannot load .env file: %v\n", err)
	}

	serverIP := dotenv.Get("SERVER_IP")
	server, err := mc.NewServer(serverIP)
	if err != nil {
		log.Fatalf("Cannot create server: %s\n", err)
	}

	client := mc.NewClient(mc.Version1_20_1)
	if err := client.Connect(server); err != nil {
		log.Fatalf("cannot connect to server: %v\n", err)
	}
	defer client.Close()

	go client.HandleResponses()

	err = client.Login("Test", "00000000-0000-4000-0000-000000000000")
	if err != nil {
		log.Fatalf("cannot login: %s", err)
	}

	select {}
}
