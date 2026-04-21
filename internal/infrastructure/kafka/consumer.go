package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/biangacila/email-service/internal/application/services"
	"github.com/biangacila/email-service/internal/domain"
	"github.com/biangacila/email-service/pkg/config"
	"github.com/biangacila/luvungula-go/global"
	"github.com/segmentio/kafka-go"
)

func StartConsumer(service *services.EmailService, cfg config.Config) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokers},
		Topic:   "emails",
		GroupID: cfg.KafkaGroupID,
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("Kafka error:", err)
			continue
		}

		var email domain.Email
		json.Unmarshal(m.Value, &email)

		global.DisplayObject(":)consumer email", email)

		service.ProcessEmail(context.Background(), email)
	}
}
