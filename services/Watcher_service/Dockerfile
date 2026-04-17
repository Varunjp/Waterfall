FROM golang:1.25.5-alpine AS builder 

WORKDIR /app 

RUN apk add --no-cache git 

COPY go.mod go.sum ./
RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o watcher-service ./cmd/worker

FROM alpine:latest 

WORKDIR /root/

COPY --from=builder /app/watcher-service .

CMD ["./watcher-service"]