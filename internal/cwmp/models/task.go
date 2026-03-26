package models

import (
	"corteca/internal/cwmp/messages"
)

type ResultsMessage struct {
	Code	int
	Message messages.Message
}

func NewResulMessage() *ResultsMessage{
	return &ResultsMessage{Code:0, Message: nil}
}
