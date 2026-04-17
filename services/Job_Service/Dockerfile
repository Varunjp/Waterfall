FROM golang:1.25.5-alpine AS builder 

WORKDIR /app 

RUN apk add --no-cache git 

COPY go.mod go.sum ./
RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o job-service ./cmd/server

FROM alpine:latest 

WORKDIR /root/

COPY --from=builder /app/job-service .
COPY --from=builder /app/.env .

EXPOSE 50051

CMD ["./job-service"]