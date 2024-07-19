package main

import (
	"encoding/json"
	"github.com/m4hi2/chatting/pkg/message"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var DefaultTimeFormat = "20060102150405"

func main() {

	mux := http.NewServeMux()

	users := map[string]int{}
	server := &sync.Map{}

	// Sender
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		sndrAddr := r.RemoteAddr + " " + r.Header.Get("user-id")
		sndr := r.Header.Get("user-id")
		messageDTO := &message.Message{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("FROM: %s -> Err Body Read: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, ok := users[sndr]; !ok {
			users[sndr] = 1
		}

		err = json.Unmarshal(body, messageDTO)
		if err != nil {
			log.Printf("FROM: %s -> Err Unmarshal: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		messageDTO.TimeStamp = time.Now().Format(DefaultTimeFormat)

		for sender := range users {
			if sender != sndr {
				value, loaded := server.Load(sender)
				if !loaded || value == nil {
					server.Store(sender, []*message.Message{messageDTO})
				} else {
					msgs := value.([]*message.Message)
					msgs = append(msgs, messageDTO)
					server.Store(sender, msgs)
				}

			}

		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{"message": "OK"}`))
		if err != nil {
			log.Printf("FROM: %s -> Err Writing to response writer: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Printf("FROM: %s -> MSG: %v \n", sndrAddr, messageDTO)

		return
	})

	// Receiver
	mux.HandleFunc("/receive/", func(w http.ResponseWriter, r *http.Request) {
		rcvrAddr := r.RemoteAddr + " " + r.Header.Get("user-id")

		rcvr := r.Header.Get("user-id")

		if _, ok := users[rcvr]; !ok {
			users[rcvr] = 1
		}

		messages, _ := server.Load(rcvr)
		marshal, err := json.Marshal(messages)
		if err != nil {
			log.Printf("FROM: %s -> Err Marshaling: %v \n", rcvrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(marshal)
		if err != nil {
			log.Printf("FROM: %s -> Err Writing to response writer: %v \n", rcvrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		server.Store(rcvr, nil)

		w.WriteHeader(http.StatusOK)
		log.Printf("FROM: %s -> Sent Messages\n", rcvrAddr)

		return
	})

	http.ListenAndServe(":8080", mux)

}
