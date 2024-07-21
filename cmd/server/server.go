package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/m4hi2/chatting/pkg/message"
	"io"
	"log"
	"net/http"
	"os"
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
	var port string
	if len(os.Args) == 2 {
		port = os.Args[1]
	} else {
		port = "8080"
	}

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
		ipToBan := r.Header.Get("ip-to-change")

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
		ipToUnban := r.Header.Get("ip-to-change")
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

	runTLSServer(mux, port)
}
func runTLSServer(handler http.Handler, port string) {
	certPem := []byte(`-----BEGIN CERTIFICATE-----
MIID+zCCAmOgAwIBAgIQTqE19CAs0FLXg2+krfg6GTANBgkqhkiG9w0BAQsFADBX
MR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExFjAUBgNVBAsMDXJvb3RA
bXlzZXJ2ZXIxHTAbBgNVBAMMFG1rY2VydCByb290QG15c2VydmVyMB4XDTI0MDcy
MTE2NTMyNFoXDTI2MTAyMTE2NTMyNFowQTEnMCUGA1UEChMebWtjZXJ0IGRldmVs
b3BtZW50IGNlcnRpZmljYXRlMRYwFAYDVQQLDA1yb290QG15c2VydmVyMIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0oDwkuuOdb21sfiwekq9+anrGO0v
fCTT6An+HqUJKN7dtEC6GnlQLQ+WC/1x8lHf5UjbEFfIrAMKbjRRd4JdxN/+35wX
RQjyunyxxEpfxIjgz/VYOxoHgVbL0lV9pZghW4E7nck4eEKNWX2JKhM7XJ2BdycQ
5RPgVfuoVJ7IiJJY/2cUumCo4DWHltCJTa4MRltYYqqTyMoRjf8ipR5nSQtLmW0Q
GtfAo49n20ABkZvU4BMITunSjvxX46pxw/Bh5bFP9L6GQzZyi4hammBbrdeTMwex
Fvf5usYPanF1vEwi/cU/yN4u5GZwsMnnzkofvK1QOxztg0iukBTX7HzUPwIDAQAB
o1kwVzAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwHwYDVR0j
BBgwFoAUYVxaoi3AQLAZrIMCfOWwmohwm5IwDwYDVR0RBAgwBocEfwAAATANBgkq
hkiG9w0BAQsFAAOCAYEAHgQA+mVGbcLEcRjOCX+QCgrjsAhhm9QPyHG6B06SH/ne
OD8LJxs43fn/s/Tgmgf1TT8WFgAkfxPRkR8M7qKYokXABI9Kk3fy86/jA5aSHqC4
r5BPK58mtBKhzPoz9NBWrfrSCUcLEzppTmBGO+YW1ZOCCamNoI9QZ6jTmPCUWM5Y
/gaB9ndiQ6hMMWjMuONXlTvM0OikmixtGSZYwbt0semxJhgKZmZ7SKBvJuCsTgtA
DZ3m7RaFsxslHAv0FpxHOuZjsWV+uWnUUl/ezWhBd8BqtsRceK2lBATn9Hpk/9gi
Di0jGDqJ0TYRoY6HtF19c1VlKQ7iNGKXYAR6YAp2QGwlZ3WXJv0R+jay+rZwLKf+
+VOxK3EWBSN+6X1jp3YelQD0W7UH6rbMaWn6YvzDxhQrXpNzqKMrtkBfBVtAGyFR
+k4CmdfMnAavwsRpNmiNFe9AfjBSCp/yFBZPjFk9qyJoJKg34dV+KdByYTtcU5P1
/LwAepRXvW6apIsi42Lp
-----END CERTIFICATE-----
`)
	keyPem := []byte(`-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDSgPCS6451vbWx
+LB6Sr35qesY7S98JNPoCf4epQko3t20QLoaeVAtD5YL/XHyUd/lSNsQV8isAwpu
NFF3gl3E3/7fnBdFCPK6fLHESl/EiODP9Vg7GgeBVsvSVX2lmCFbgTudyTh4Qo1Z
fYkqEztcnYF3JxDlE+BV+6hUnsiIklj/ZxS6YKjgNYeW0IlNrgxGW1hiqpPIyhGN
/yKlHmdJC0uZbRAa18Cjj2fbQAGRm9TgEwhO6dKO/FfjqnHD8GHlsU/0voZDNnKL
iFqaYFut15MzB7EW9/m6xg9qcXW8TCL9xT/I3i7kZnCwyefOSh+8rVA7HO2DSK6Q
FNfsfNQ/AgMBAAECggEAPP7bttbGqttTwMQc7vKlZaiU1N41ejV5qazrk5mis9MQ
TuDKjE1GrCfuBH9l+x86T0fzIiMtpJok9ZX3XTfLT/bP9Z9XJsvW+a6UHBqo8Vvw
OJIRBN8f+Zxa5xGanNceI1OpIKlj0YUHTD0R970m4ElLcGlDff1qbb/EIPD5jojC
nvj9P4r+iYiJHI39wjmxrHjW53LczbC3CdQIkA77Q70wFzwxVabqKtRHCpMKgtN/
9ujBdH0IMV1whMUAM/ZaZ3bhkQ90eqCn+1Fompcr9eLqLWwmuUCuyGkUo2gbO+qv
0OAbQjm6QzTyT0KbC3i+IQVR+01UCp3PRIFB9T8+EQKBgQD5IVqs/cVlvnNE9KwZ
Ag8Y7iX7rR4cmh6my3dY6g0O8OFFi45IOOomrMlkqRapS00BNMeUXYI1WWUqJWZu
h0e3bmoRWgNivivMg9EXO+fctV1bnq0VFd8CoNfMvgazaN6LlBXT/ltY8xhc79zl
aHhAC3LDJ1H6QrjC54kMSUGHOwKBgQDYTuowIoSWI6+DaESZfV9SmPaTu9xgT7sg
NC+Xa1JU4jOEM5zTPEXgQZX3/tNZiyrnQ8CtRgDoiNpQhVRPr9usFgmfnOawGJ0m
Qnkr92ZRA4v74c9gwDWEXXij1h3UQUi2Lo9BDXF+K414ZfjeS4/nfMTZsZTELZVQ
fYkKBRX+zQKBgQDzXByqjf8NA3ywaF3Q1A0RallaP2MBx5+XiXwdNAzrgmxcNhYY
ANjiTLkyhmYnm/It8nPfP3TZTmkfQYeNAsMQsWAVcRuLkn9QeZd/nbpCBDydKiSj
S0kc1SfYzevXx+JU8KReAMMU4erzpi/fPSzySvyhHEDdOd4oLmrWwJytTwKBgDXn
lJziPUBdLEQHG/FUOQkQbYJrcoPd2rgvyRxm9mGI7WWopxBNGOzF62Wd73WFJO/1
BnMn2toYIK+oSlaBcLD13PgV1bdUqvT549B9GtZdl+jxYQivXaba1FGf7lcS3dfo
ynJfs0TJ/btfTiG3mukJQtUtV/F7mpYwcpI4qj6RAoGBALvNt3+HXYuPwb/56gxW
HMXUijNCozyZF95S9DPcoK85WgL70r3C2k6gRaqmf4Rp1crSaTfb0Cuh3cHJatzT
SUGnjqMLP5OqoZtK9NbhdVV57Ens3uag3dgF+cmmn08EzCGAec9Z8SrAW1dgSb1Z
ryYWM0q9spwKAMZVbP1wznPo
-----END PRIVATE KEY-----
`)
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	srv := &http.Server{
		TLSConfig:    cfg,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		Addr:         addr,
		Handler:      handler,
	}
	log.Fatal(srv.ListenAndServeTLS("", ""))

}
