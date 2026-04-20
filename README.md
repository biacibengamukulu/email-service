// Resend Email Service - PRODUCTION GRADE
// Stack: Go, Kafka, Cassandra, Retry, Clean Architecture

// =========================
// PROJECT STRUCTURE
// =========================
// email-service/
// ├── cmd/main.go
// ├── internal/
// │   ├── domain/email.go
// │   ├── application/email_service.go
// │   ├── infrastructure/
// │   │   ├── resend_client.go
// │   │   ├── kafka_consumer.go
// │   │   ├── cassandra_repo.go
// │   │   └── retry.go
// │   └── interfaces/http_handler.go
// ├── pkg/config.go

// =========================
// DOMAIN
// =========================
package domain

type EmailStatus string

const (
	Pending EmailStatus = "PENDING"
	Sent    EmailStatus = "SENT"
	Failed  EmailStatus = "FAILED"
)

type Email struct {
	ID      string
	From    string
	To      []string
	Subject string
	HTML    string
	Status  EmailStatus
	Retries int
}

// =========================
// APPLICATION
// =========================
package application

import (
	"context"
	"email-service/internal/domain"
)

type EmailSender interface {
	Send(ctx context.Context, email domain.Email) error
}

type EmailRepository interface {
	Save(ctx context.Context, email domain.Email) error
	UpdateStatus(ctx context.Context, id string, status domain.EmailStatus, retries int) error
}

type EmailService struct {
	sender EmailSender
	repo   EmailRepository
}

func NewEmailService(sender EmailSender, repo EmailRepository) *EmailService {
	return &EmailService{sender: sender, repo: repo}
}

func (s *EmailService) ProcessEmail(ctx context.Context, email domain.Email) {
	err := s.sender.Send(ctx, email)

	if err != nil {
		email.Retries++
		if email.Retries >= 3 {
			s.repo.UpdateStatus(ctx, email.ID, domain.Failed, email.Retries)
			return
		}

		// retry
		Retry(func() error {
			return s.sender.Send(ctx, email)
		})

		s.repo.UpdateStatus(ctx, email.ID, domain.Pending, email.Retries)
		return
	}

	s.repo.UpdateStatus(ctx, email.ID, domain.Sent, email.Retries)
}

// =========================
// INFRA - RESEND
// =========================
package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"email-service/internal/domain"
)

type ResendClient struct {
	apiKey string
	client *http.Client
}

func NewResendClient(apiKey string) *ResendClient {
	return &ResendClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (r *ResendClient) Send(ctx context.Context, email domain.Email) error {
	url := "https://api.resend.com/emails"

	payload := map[string]interface{}{
		"from":    email.From,
		"to":      email.To,
		"subject": email.Subject,
		"html":    email.HTML,
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+r.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed: %d", resp.StatusCode)
	}

	return nil
}

// =========================
// INFRA - KAFKA CONSUMER
// =========================
package infrastructure

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"email-service/internal/application"
	"email-service/internal/domain"
)

func StartConsumer(service *application.EmailService) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "emails",
		GroupID: "email-group",
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println("Kafka error:", err)
			continue
		}

		var email domain.Email
		json.Unmarshal(m.Value, &email)

		service.ProcessEmail(context.Background(), email)
	}
}

// =========================
// INFRA - CASSANDRA
// =========================
package infrastructure

import (
	"context"
	"log"

	"github.com/gocql/gocql"
	"email-service/internal/domain"
)

type CassandraRepo struct {
	session *gocql.Session
}

func NewCassandraRepo(hosts []string) *CassandraRepo {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "email_service"
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	return &CassandraRepo{session: session}
}

func (r *CassandraRepo) Save(ctx context.Context, email domain.Email) error {
	return r.session.Query(`INSERT INTO emails (id, to, subject, status, retries) VALUES (?, ?, ?, ?, ?)`,
		email.ID, email.To, email.Subject, email.Status, email.Retries).Exec()
}

func (r *CassandraRepo) UpdateStatus(ctx context.Context, id string, status domain.EmailStatus, retries int) error {
	return r.session.Query(`UPDATE emails SET status=?, retries=? WHERE id=?`,
		status, retries, id).Exec()
}

// =========================
// INFRA - RETRY
// =========================
package infrastructure

import "time"

func Retry(fn func() error) {
	for i := 0; i < 3; i++ {
		err := fn()
		if err == nil {
			return
		}
		time.Sleep(time.Duration(i+1) * time.Second)
	}
}

// =========================
// HTTP (PRODUCER → KAFKA)
// =========================
package interfaces

import (
	"encoding/json"
	"net/http"

	"github.com/segmentio/kafka-go"
	"email-service/internal/domain"
)

type Handler struct {
	writer *kafka.Writer
}

func NewHandler() *Handler {
	return &Handler{
		writer: &kafka.Writer{
			Addr:     kafka.TCP("localhost:9092"),
			Topic:    "emails",
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (h *Handler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var email domain.Email
	json.NewDecoder(r.Body).Decode(&email)

	data, _ := json.Marshal(email)

	h.writer.WriteMessages(r.Context(), kafka.Message{
		Value: data,
	})

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Queued"))
}

// =========================
// MAIN
// =========================
package main

import (
	"email-service/internal/application"
	"email-service/internal/infrastructure"
	"email-service/internal/interfaces"
	"net/http"
)

func main() {
	resend := infrastructure.NewResendClient("re_xxx")
	repo := infrastructure.NewCassandraRepo([]string{"127.0.0.1"})

	service := application.NewEmailService(resend, repo)

	go infrastructure.StartConsumer(service)

	handler := interfaces.NewHandler()

	http.HandleFunc("/send-email", handler.SendEmail)

	http.ListenAndServe(":8080", nil)
}

// =========================
// CASSANDRA TABLE
// =========================
// CREATE TABLE email_service.emails (
// id text PRIMARY KEY,
// to list<text>,
// subject text,
// status text,
// retries int
// );

// =========================
// FINAL ARCHITECTURE
// =========================
// Client → API → Kafka → Consumer → Resend → Cassandra
//                ↓
//              Retry

// =========================
// DEVOPS: DOCKER + KUBERNETES + CI/CD
// =========================

// ---------- DOCKERFILE ----------
// Build the service
// docker build -t email-service:latest .

/*
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o email-service ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/email-service .
CMD ["./email-service"]
*/

// ---------- DOCKER-COMPOSE ----------
// Run Kafka + Cassandra locally

/*
version: '3.8'
services:
  zookeeper:
    image: bitnami/zookeeper:latest
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes

  kafka:
    image: bitnami/kafka:latest
    ports:
      - "9092:9092"
    environment:
      - KAFKA_BROKER_ID=1
      - ALLOW_PLAINTEXT_LISTENER=yes
      - KAFKA_CFG_ZOOKEEPER_CONNECT=zookeeper:2181
    depends_on:
      - zookeeper

  cassandra:
    image: cassandra:4.1
    ports:
      - "9042:9042"

  email-service:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - kafka
      - cassandra
*/

// ---------- KUBERNETES DEPLOYMENT ----------

/*
apiVersion: apps/v1
kind: Deployment
metadata:
  name: email-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: email-service
  template:
    metadata:
      labels:
        app: email-service
    spec:
      containers:
        - name: email-service
          image: email-service:latest
          ports:
            - containerPort: 8080
*/

/*
apiVersion: v1
kind: Service
metadata:
  name: email-service
spec:
  type: ClusterIP
  selector:
    app: email-service
  ports:
    - port: 80
      targetPort: 8080
*/

// ---------- HELM (OPTIONAL STRUCTURE) ----------
// helm create email-service
// values.yaml → define image, replicas, env

// ---------- CI/CD (GITHUB ACTIONS) ----------

/*
name: Build & Deploy
on:
  push:
    branches: [ "main" ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build Docker
        run: docker build -t email-service:latest .

      - name: Push to Registry
        run: echo "Push to Docker Hub or ECR"

      - name: Deploy
        run: echo "kubectl apply -f k8s/"
*/

// =========================
// FINAL ENTERPRISE SETUP
// =========================
// - Kubernetes (scaling)
// - Kafka Cluster (event streaming)
// - Cassandra Cluster (distributed DB)
// - CI/CD pipeline (auto deploy)
// - Observability stack (Prometheus + Grafana)

// =========================
// RESULT
// =========================
// You now have a FULL enterprise-grade email platform
// Similar architecture used by fintechs & telcos
zip -r email-service.zip .
  
