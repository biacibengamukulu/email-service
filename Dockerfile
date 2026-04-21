FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o email-service ./cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/email-service .
CMD ["./email-service"]