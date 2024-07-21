package main

import (
	"encoding/json"
	"fmt"
	"github.com/m4hi2/chatting/pkg/message"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const ADMINCREDS = "bablaanis"

var DefaultTimeFormat = "20060102150405"

type UserData struct {
	IP       string    `json:"ip"`
	OnlineAt time.Time `json:"online_at"`
}

func main() {

	mux := http.NewServeMux()

	banned := map[string]bool{}
	users := map[string]*UserData{}
	server := &sync.Map{}

	// Sender
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		sndrAddr := r.RemoteAddr + " " + r.Header.Get("user-id")
		sndrIP := strings.Split(r.RemoteAddr, ":")[0]
		sndr := r.Header.Get("user-id")
		messageDTO := &message.Message{}

		w.Header().Set("Content-Type", "application/json")

		if bn := banned[sndrIP]; bn {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"banned"}`))
			log.Printf("banned: %s", sndrIP)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("FROM: %s -> Err Body Read: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, ok := users[sndr]; !ok {
			users[sndr] = &UserData{
				IP:       sndrIP,
				OnlineAt: time.Now(),
			}
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

		_, err = w.Write([]byte(`{"message": "OK"}`))
		if err != nil {
			log.Printf("FROM: %s -> Err Writing to response writer: %v \n", sndrAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	})

	// Receiver
	mux.HandleFunc("/receive/", func(w http.ResponseWriter, r *http.Request) {
		rcvrAddr := r.RemoteAddr + " " + r.Header.Get("user-id")
		rcvrIP := strings.Split(r.RemoteAddr, ":")[0]

		w.Header().Set("Content-Type", "application/json")
		if bn := banned[rcvrIP]; bn {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"banned"}`))
			log.Printf("banned: %s", rcvrIP)
			return
		}

		rcvr := r.Header.Get("user-id")

		users[rcvr] = &UserData{
			IP:       rcvrIP,
			OnlineAt: time.Now(),
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

	mux.HandleFunc("/user-status", func(w http.ResponseWriter, r *http.Request) {
		adminID := r.Header.Get("admin-id")
		w.Header().Set("Content-Type", "application/json")
		if adminID != ADMINCREDS {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		data, err := json.Marshal(users)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("FROM: %s -> Err Marshaling (User Stats): %v \n", adminID, err)
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			log.Printf("FROM: %s -> Err Writing to response (User Stats): %v \n", adminID, err)
		}

	})

	mux.HandleFunc("/ban", func(w http.ResponseWriter, r *http.Request) {
		adminID := r.Header.Get("admin-id")
		ipToBan := r.Header.Get("ip-to-ban")

		w.Header().Set("Content-Type", "application/json")
		if adminID != ADMINCREDS {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		banned[ipToBan] = true

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf(`{"message": "OK - Banned IP: %s"}`, ipToBan)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("FROM: %s -> Err Writing to response writer (Ban): %v \n", adminID, err)
			return
		}
	})

	mux.HandleFunc("/show-banned", func(w http.ResponseWriter, r *http.Request) {
		adminID := r.Header.Get("admin-id")

		w.Header().Set("Content-Type", "application/json")
		if adminID != ADMINCREDS {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		data, err := json.Marshal(banned)
		if err != nil {
			log.Printf("FROM: %s -> Err Marshaling (Show Banned): %v \n", adminID, err)
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			log.Printf("FROM: %s -> Err Writing to response writer (Show Banned): %v \n", adminID, err)
			return
		}
	})

	mux.HandleFunc("/unban", func(w http.ResponseWriter, r *http.Request) {
		adminID := r.Header.Get("admin-id")
		ipToUnban := r.Header.Get("ip-to-unban")
		w.Header().Set("Content-Type", "application/json")

		if adminID != ADMINCREDS {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		banned[ipToUnban] = false

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf(`{"message": "OK - Un Banned IP: %s"}`, ipToUnban)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("FROM: %s -> Err Writing to response writer (Un Ban): %v \n", adminID, err)
			return
		}

	})

	http.ListenAndServe("0.0.0.0:8080", mux)

}
