FROM golang:1.25.5-alpine AS builder

WORKDIR /app 

RUN apk add --no-cache git 

COPY go.mod go.sum ./
RUN go mod download 

COPY . . 

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o admin-service ./cmd/server 

FROM alpine:latest 

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/admin-service .
COPY --from=builder /app/.env .

EXPOSE 8080 50054 9101

CMD ["./admin-service"]
