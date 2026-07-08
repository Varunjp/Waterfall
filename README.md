# Waterfall — Multi-Tenant SaaS Distributed Job Scheduling Platform

A production-grade job scheduling platform built with **Go, PostgreSQL, Redis, and Kafka**, designed around **Clean Architecture** and **Domain-Driven Design**.

The platform allows tenants to register applications, schedule asynchronous jobs (immediate or deferred), monitor execution, and manage usage-based billing — while giving platform administrators full operational control and observability.

---

## Problem Statement

Modern SaaS systems require:

- Reliable background job execution
- Scheduled and delayed job processing
- Multi-tenant isolation
- Retry handling with failure tracking
- Usage-based billing
- Observability and monitoring
- Secure authentication and authorization

This platform addresses these requirements in a scalable, production-ready architecture.

---

## Architecture Overview

The system is built on:

- **Clean Architecture** — clear separation between domain, use case, and delivery layers
- **Domain-Driven Design** — business logic modeled around entities, aggregates, and domain services
- **Repository Pattern** — a persistence abstraction used within the domain/infrastructure boundary, in line with standard DDD tactical design.
- **SOLID Principles**
- **Microservices Architecture** — each service is independently deployable and scalable

---

## Microservices

1. **API Service** — Public REST APIs for tenants
2. **Admin Service** — Platform administration, tenant management, and billing
3. **Job Service** — Core job management and state machine
4. **Scheduler Service** — Moves due scheduled jobs to the queue
5. **Watcher Service** — Monitors worker execution and persists job outcomes to the database

---

## Tech Stack

**Backend**
- Golang
- PostgreSQL
- Redis Streams / Kafka
- JWT Authentication
- RBAC Authorization
- Prometheus
- Grafana

**Infrastructure**
- Docker
- Docker Compose
- Environment-based configuration (.env)
- Structured logging

---

## Clean Architecture Structure

Each service follows this structure:

```text
├── domain/
│   ├── entity/
│   │   └── job.go
│   ├── repository/
│   │   └── job_repository.go
│   └── service/
│       └── job_service.go
│
├── usecase/
│   └── job_usecase.go
│
├── repository/
│   └── postgres_job_repository.go
│
├── delivery/
│   ├── http/
│   │   └── job_handler.go
│   └── grpc/
│       └── job_grpc.go
│
└── dto/
    ├── create_job_request.go
    └── job_response.go
```

**Architecture flow:**
`Controller → UseCase → Repository Interface → Repository Implementation`

This ensures:
- Testability
- Loose coupling
- Dependency inversion
- Maintainability

---

## Authentication & Authorization

**Platform Admin**
- Admin login
- JWT access + refresh tokens
- List all tenants
- Block / unblock tenants
- View billing per tenant
- Inspect job execution and status

**Tenant**
- Tenant registration
- AppID and API key generation
- Worker registration
- RBAC user roles
- Tenant login (JWT)
- Monitor jobs by role

---

## Job Lifecycle

**Immediate Jobs**
1. Tenant creates a job via API
2. Job is stored in PostgreSQL
3. Job is pushed to Redis/Kafka
4. Worker consumes and executes the job
5. Status is updated

**Scheduled Jobs**
1. `schedule_at` column is used
2. Job state is set to `SCHEDULED`
3. Scheduler scans for due jobs
4. Job is moved to the queue
5. Worker executes the job

---

## Reliability Features

- Retry mechanism with exponential backoff
- `retry_count` and `max_retries` tracking
- Job state machine enforcement
- Idempotent execution
- Execution locks
- Correlation IDs for tracing

---

## Performance: Kafka Writer Tune

![Kafka_tune benchmark](docs/performance_tune_chart.svg)

**Full load test reports**

- [k6 report — before tuning](docs/reports/report_before.html)
- [k6 report — after tuning](docs/reports/report.html)

---

## Billing & Monetization

*(Handled by the Admin Service)*

- Tier-based pricing
- Per-job execution tracking
- Per-retry tracking
- Daily usage metering
- Subscription plans
- Invoice generation
- Quota enforcement
- Soft limits and grace periods
- Automatic job blocking when quota is exceeded

---

## Observability

**Prometheus Metrics**
- Jobs created
- Jobs succeeded
- Jobs failed
- Retry count
- Queue depth
- Worker count
- Throughput

**Grafana Dashboards**
- Job throughput dashboard
- Service health dashboard
- Queue lag monitoring
- Failure rate visualization

---

## Database Tables

Core tables:

- `platform_admins`
- `app_users`
- `apps`
- `jobs`
- `job_logs`
- `usage_daily`
- `subscriptions`
- `payments`

---

## Job State Machine

```
CREATED → QUEUED → RUNNING → SUCCESS
                       ↓
                    RETRYING → FAILED
```

- `RETRYING` is entered on execution failure, up to `max_retries` attempts.
- Once retries are exhausted, the job transitions to a terminal `FAILED` state.
- Only valid transitions defined above are permitted.

---

## Security

- JWT-based authentication
- Role-based access control (RBAC)
- API key authentication for job creation
- Middleware-based token validation
- Tenant isolation
- Environment-based secret management
- No hardcoded credentials

---

## Example APIs

**Tenant**

```
POST /api/v1/apps
POST /api/v1/users/login
POST /api/v1/users
POST /api/v1/jobs 
GET  /api/v1/jobs/{job_id}/logs
```

**Admin**

```
POST  /api/v1/admin/login
GET   /api/v1/apps
PATCH /api/v1/apps/{app_id}/block
GET   /api/v1/admin/payments
GET   /api/v1/admin/plans
```

---

## Running the Project

```bash
# Clone repository
git clone https://github.com/Varunjp/waterfall.git
cd waterfall

# Set up environment variables
cp .env.example .env

# Install dependencies
go mod tidy

# Run all services
go run main.go
```

**.env file**

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=waterfall

REDIS_HOST=localhost
REDIS_PORT=6379

KAFKA_BROKER=localhost:9092

JWT_SECRET=supersecret
```

---

## Testing Strategy

- Unit tests for use cases
- Repository mock testing
- Integration testing
- Retry simulation tests
- Scheduler restart tests
- Load testing for scaling

---

## Engineering Decisions

| Decision | Rationale |
|---|---|
| Clean Architecture | Maintainability and testability |
| Redis Streams / Kafka | Distributed queue reliability |
| PostgreSQL | ACID guarantees |
| JWT | Stateless authentication |
| RBAC | Multi-role security |
| Microservices | Independent scaling |
| Prometheus | Native Go support |
| Structured logging | Production debugging |

---

## Learning Outcomes

This project demonstrates:

- Distributed systems fundamentals
- Atleast-once execution simulation
- Retry and failure-handling design patterns
- Multi-tenant SaaS architecture
- Billing system design
- Production monitoring setup
- Clean Architecture and DDD in practice
- System reliability engineering

---

## Production Readiness Checklist

- [x] Structured logging
- [x] Centralized error handling
- [x] JWT authentication
- [x] RBAC
- [x] Retry engine
- [x] Scheduler reliability
- [x] Billing enforcement
- [x] Metrics and monitoring
- [x] Dockerized services
- [x] Environment-based configuration

---

## Author

**Varun JP**
Backend Engineer — Golang
Distributed Systems & Scalable Architecture Enthusiast
