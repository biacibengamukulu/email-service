
# Deploy
cd deploy

# Build
docker compose build

# Push
docker compose push

# Update live server
cd deploy && \
docker compose build && \
docker compose push && \
ssh safer "docker pull 010309/email-service:latest && \
docker rm -f email-service && \
cd /apps/docker-compose-script/swalydelivery && docker-compose up -d email-service"

