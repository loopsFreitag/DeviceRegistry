### BUILDER STAGE ###
FROM golang:1.24-alpine AS builder
RUN apk update && \
  apk add --no-cache make git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.12

WORKDIR /src
# Copy Go dependency definitions separately to take advantage of build caching
COPY go.mod .
COPY go.sum .
ENV GOOS=linux
ENV GOARCH=amd64
RUN go mod download
COPY . .
# Generate Swagger docs
RUN swag init
RUN VERSION=$(git describe --always --tags --abbrev=9 --long --match 'v[0-9]*.[0-9]*.[0-9]*' 2>/dev/null || echo "v0.0.0-dev") && \
  make build

### BASE STAGE ###
FROM alpine:3.17.2 AS base
RUN apk update && \
  apk add --no-cache ca-certificates netcat-openbsd
EXPOSE 8081
WORKDIR /app
COPY --from=builder /src/build/* ./
COPY --from=builder /src/config*.yml ./
COPY --from=builder /src/entrypoint.sh .
COPY --from=builder /go/bin/goose .
# Copy generated swagger docs
COPY --from=builder /src/docs ./docs
# DB migration files
COPY --from=builder /src/db/migrations ./db/migrations
# Make scripts executable
RUN chmod +x ./entrypoint.sh ./deviceregistry ./goose

### DEV STAGE ###
FROM base AS dev
ENTRYPOINT ["./entrypoint.sh"]

### PROD STAGE ###
FROM base AS prod
RUN addgroup -S deviceregistry && adduser -S deviceregistry -G deviceregistry
USER deviceregistry
ENTRYPOINT ["./entrypoint.sh"]
