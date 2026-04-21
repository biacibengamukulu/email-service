package e2e_test

import (
	"log"
	"testing"

	"github.com/biangacila/email-service/internal/application/services"
	"github.com/biangacila/email-service/pkg/config"
	"github.com/joho/godotenv"
)

func TestSmsService(t *testing.T) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env")
	}

	cfg := config.Load("")

	log.Println(cfg)
	t.Run("Send", func(t *testing.T) {
		svc := services.NewSmsService(&cfg)
		if svc == nil {
			t.Fatal("Failed to create sms service")
		}

		err := svc.Send("27729139504", "KZN sms Gateway Test message")
		if err != nil {
			t.Errorf("Failed to send sms: %v", err)
		}

	})
}
