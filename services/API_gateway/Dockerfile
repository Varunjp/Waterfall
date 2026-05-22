FROM golang:1.25.5-alpine AS builder 

WORKDIR /app 

RUN apk add --no-cache git 

COPY go.mod go.sum ./
RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o api-gateway ./cmd/main.go 

FROM alpine:latest 

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/api-gateway . 

COPY --from=builder /app/web ./web 

EXPOSE 8081

CMD ["./api-gateway"]