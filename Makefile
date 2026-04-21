SHELL := /bin/bash

run-email-service:
	 go run ./cmd/api

run-deploy:
	cd deploy && \
    docker compose build && \
    docker compose push && \
    ssh safer "docker pull 010309/email-service:latest && \
    docker rm -f email-service && \
    cd /apps/docker-compose-script/swaly && docker-compose up -d email-service"
