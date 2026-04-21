package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/biangacila/email-service/internal/application/services"
	"github.com/biangacila/email-service/internal/infrastructure/cassandra"
	"github.com/biangacila/email-service/internal/infrastructure/kafka"
	"github.com/biangacila/email-service/internal/infrastructure/resend"
	"github.com/biangacila/email-service/internal/interfaces/handlers"
	"github.com/biangacila/email-service/internal/migrations"
	"github.com/biangacila/email-service/pkg/config"
	"github.com/biangacila/email-service/pkg/logger"
	"github.com/biangacila/luvungula-go/global"
	"github.com/joho/godotenv"
)

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func parseInt(s string, def int) int {
	n := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return def
	}
	return n
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env")
	}

	cfg := config.Load("biatech-email-service")
	global.DisplayObject("config", cfg)
	resendSvg := resend.NewResendClient(cfg.ResendToken)
	repo := cassandra.NewCassandraRepo([]string{cfg.CassandraHosts})

	// let do migration
	log := logger.New(cfg.ServiceName)
	rf := parseInt(config.Getenv("CASSANDRA_RF", "1"), 1)
	if err := cassandra.BootstrapKeyspaceAndTables(cfg.CassandraHosts, cfg.CassandraKeyspace, rf, migrations.DDL); err != nil {
		log.Fatalf("cassandra bootstrap: %v", err)
	}

	service := services.NewEmailService(resendSvg, repo)
	svcSms := services.NewSmsService(&cfg)

	go kafka.StartConsumer(service, cfg)

	handlerEmail := handlers.NewHandler(cfg)
	handlerSms := handlers.NewSmsHandler(svcSms)

	prefix := "/backend-email-service/api/v1"

	http.HandleFunc(prefix+"/send-email", handlerEmail.SendEmail)
	http.HandleFunc(prefix+"/send-sms/post", handlerSms.SendPost)
	http.HandleFunc(prefix+"/send-sms/get", handlerSms.SendGet)

	http.ListenAndServe(":"+cfg.HTTPPort, nil)
}
