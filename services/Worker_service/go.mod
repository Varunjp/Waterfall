module worker_service

go 1.25.5

require (
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.10
)

require go.uber.org/multierr v1.10.0 // indirect

require (
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.27.1
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
)
