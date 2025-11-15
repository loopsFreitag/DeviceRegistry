### BUILDER STAGE ###
FROM golang:1.23-alpine AS builder

RUN apk update && \
  apk add --no-cache make git

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /src

# Copy Go dependency definitions separately to take advantage of build caching
COPY go.mod .
COPY go.sum .

ENV GOOS=linux
ENV GOARCH=amd64

RUN go mod download

COPY . .

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
