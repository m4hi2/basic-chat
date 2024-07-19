package main

import (
	"github.com/m4hi2/chatting/pkg/message"
	"log"
	"testing"
)

func TestReceive(t *testing.T) {
	msg := CollectMessages("mahir", "127.0.0.1", "8080")
	for _, m := range msg {
		log.Println(m)
	}
}

func TestSend(t *testing.T) {
	msg := &message.Message{
		From:    "tahmid",
		Message: "gg test",
	}
	err := SendMessage("tahmid", "127.0.0.1", "8080", msg)

	if err != nil {
		log.Fatal(err)
	}
}
