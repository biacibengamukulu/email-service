package main

import (
	"context"
	"encoding/json"
	"errors"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/biangacila/email-service/internal/application/services"
	"github.com/biangacila/email-service/internal/infrastructure/cassandra"
	"github.com/biangacila/email-service/internal/infrastructure/kafka"
	"github.com/biangacila/email-service/internal/infrastructure/resend"
	"github.com/biangacila/email-service/internal/interfaces/handlers"
	"github.com/biangacila/email-service/internal/migrations"
	"github.com/biangacila/email-service/pkg/config"
	"github.com/biangacila/email-service/pkg/logger"
	"github.com/joho/godotenv"
)

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		stdlog.Printf("json marshal error: %v", err)
		return []byte("{}")
	}
	return b
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return def
		}
		n = n*10 + int(s[i]-'0')
	}
	if n <= 0 {
		return def
	}
	return n
}

func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		stdlog.Println("No .env file found, using system env")
	}

	cfg := config.Load("biatech-email-service")

	log := logger.New(cfg.ServiceName)

	// Init infrastructure
	resendClient := resend.NewResendClient(cfg.ResendToken)
	repo := cassandra.NewCassandraRepo([]string{cfg.CassandraHosts})

	// Cassandra bootstrap
	rf := parseInt(config.Getenv("CASSANDRA_RF", "1"), 1)
	if err := cassandra.BootstrapKeyspaceAndTables(
		cfg.CassandraHosts,
		cfg.CassandraKeyspace,
		rf,
		migrations.DDL,
	); err != nil {
		log.Fatalf("cassandra bootstrap failed: %v", err)
	}

	// Services
	emailService := services.NewEmailService(resendClient, repo)
	smsService := services.NewSmsService(&cfg)

	// Kafka consumer (with context)
	_, cancel := context.WithCancel(context.Background())
	go func() {
		log.Println("Starting Kafka consumer...")
		kafka.StartConsumer(emailService, cfg)
	}()

	// Handlers
	handlerEmail := handlers.NewHandler(cfg)
	handlerSms := handlers.NewSmsHandler(smsService)

	// Router (better than default mux)
	mux := http.NewServeMux()

	prefix := "/backend-email-service/api/v1"

	mux.HandleFunc(prefix+"/send-email", handlerEmail.SendEmail)
	mux.HandleFunc(prefix+"/send-sms/post", handlerSms.SendPost)
	mux.HandleFunc(prefix+"/send-sms/get", handlerSms.SendGet)

	// HTTP Server with timeouts
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Println("HTTP server running on port " + cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	// Stop Kafka
	cancel()

	// Shutdown HTTP server
	ctxShutdown, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	log.Println("Server exited cleanly")
}
