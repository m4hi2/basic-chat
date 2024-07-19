package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/m4hi2/chatting/pkg/message"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	user := os.Args[1]
	serverIP := os.Args[2]
	serverPort := os.Args[3]

	go func() {
		for {
			time.Sleep(1 * time.Second)
			messages := CollectMessages(user, serverIP, serverPort)
			for _, msg := range messages {
				log.Printf("%s 🗣 ️%s", msg.From, msg.Message)
			}
		}
	}()

	msgs := []string{}
	for {
		var msg string
		_, err := fmt.Scan(&msg)
		if err != nil {
			log.Fatal(err)
		}
		msgs = append(msgs, msg)
		if strings.Contains(msg, ";") {
			msgLine := strings.Join(msgs, " ")
			msgs = nil
			newmsg := strings.Trim(msgLine, ";")
			msgdto := &message.Message{
				From:    user,
				Message: newmsg,
			}
			err := SendMessage(user, serverIP, serverPort, msgdto)
			if err != nil {
				log.Println("Error sending message: ", err)
				return
			}

			log.Printf("%s 🗣️ %s", user, newmsg)

		}
	}

}

func CollectMessages(user string, host string, port string) []*message.Message {
	messages := []*message.Message{}
	url := fmt.Sprintf("http://%s:%s/receive", host, port)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", user)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil
	}

	err = json.Unmarshal(responseBody, &messages)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil
	}

	return messages
}

func SendMessage(user string, host string, port string, message *message.Message) error {
	url := fmt.Sprintf("http://%s:%s/send", host, port)
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	br := bytes.NewReader(body)

	req, err := http.NewRequest("POST", url, br)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-id", user)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("error: %v", response.Status)
	}

	return nil
}
