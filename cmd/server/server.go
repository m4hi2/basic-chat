package main

import (
	"encoding/json"
	"github.com/m4hi2/chatting/pkg/message"
	"io"
	"log"
	"net/http"
	"time"
)

var DefaultTimeFormat = "20060102150405"

func main() {

	mux := http.NewServeMux()

	msgFromMahir := []*message.Message{}
	msgFromTahmid := []*message.Message{}

	// Sender
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		sndrAddr := r.RemoteAddr
		message := &message.Message{}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("FROM: %s -> Err Body Read: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(body, message)
		if err != nil {
			log.Printf("FROM: %s -> Err Unmarshal: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		message.TimeStamp = time.Now().Format(DefaultTimeFormat)

		if message.From == "tahmid" {
			msgFromTahmid = append(msgFromTahmid, message)
		} else if message.From == "mahir" {
			msgFromMahir = append(msgFromMahir, message)
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{"message": "OK"}`))
		if err != nil {
			log.Printf("FROM: %s -> Err Writing to response writer: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		log.Printf("FROM: %s -> MSG: %v \n", sndrAddr, message)

		return
	})

	// Receiver
	mux.HandleFunc("/receive/", func(w http.ResponseWriter, r *http.Request) {
		rcvrAddr := r.RemoteAddr

		rcvr := r.Header.Get("user-id")
		if rcvr == "mahir" {
			marshal, err := json.Marshal(msgFromTahmid)
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

			w.WriteHeader(http.StatusOK)
			msgFromTahmid = nil
		} else if rcvr == "tahmid" {
			marshal, err := json.Marshal(msgFromMahir)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Printf("FROM: %s -> Err Marshaling: %v \n", rcvrAddr, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(marshal)
			if err != nil {
				log.Printf("FROM: %s -> Err Writing to response writer: %v \n", rcvrAddr, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			msgFromMahir = nil
		}

		log.Printf("FROM: %s -> Sent Messages\n", rcvrAddr)

	})

	http.ListenAndServe(":8080", mux)

}
