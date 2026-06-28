FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o entropix ./cmd/gateway

FROM alpine:3.21

RUN apk add --no-cache ca-certificates
COPY --from=builder /app/entropix /usr/local/bin/entropix
COPY --from=builder /app/config.yaml /etc/entropix/config.yaml

EXPOSE 8080

ENTRYPOINT ["entropix"]
CMD ["--config", "/etc/entropix/config.yaml"]
