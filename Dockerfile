FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o secrets-vault ./cmd/api

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/secrets-vault .

COPY .env /app/.env

CMD ["./secrets-vault"]
