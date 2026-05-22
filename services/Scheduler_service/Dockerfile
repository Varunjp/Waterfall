FROM golang:1.25.5-alpine AS builder 

WORKDIR /app 

RUN apk add --no-cache git 

COPY go.mod go.sum ./
RUN go mod download 

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o scheduler-service ./cmd/server

FROM alpine:latest 

WORKDIR /app  

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/scheduler-service .

COPY --from=builder /app/internal/redis/lua ./internal/redis/lua 

EXPOSE 50052 9100

CMD ["./scheduler-service"]