FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN go build -o qa-service ./main.go

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/qa-service /app/qa-service
COPY db/migrations /app/db/migrations

ENV DATABASE_DSN="host=db user=postgres password=postgres dbname=qa_service port=5432 sslmode=disable"

EXPOSE 8080

CMD goose -dir /app/db/migrations postgres "$DATABASE_DSN" up && /app/qa-service



