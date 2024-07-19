package message

type Message struct {
	From      string `json:"from"`
	Message   string `json:"message"`
	TimeStamp string `json:"timestamp"`
}
