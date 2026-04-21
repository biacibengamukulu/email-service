package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/biangacila/email-service/internal/domain"
	"github.com/biangacila/email-service/pkg/config"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type EmailHandler struct {
	writer *kafka.Writer
}

func NewHandler(cfg config.Config) *EmailHandler {
	return &EmailHandler{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(cfg.KafkaBrokers),
			Topic:    "emails",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (h *EmailHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var email domain.Email
	json.NewDecoder(r.Body).Decode(&email)

	emailId := uuid.New().String()
	email.ID = emailId

	data, _ := json.Marshal(email)

	h.writer.WriteMessages(r.Context(), kafka.Message{
		Value: data,
	})

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Queued"))
}
