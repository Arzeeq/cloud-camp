# Stage builder
FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ratelimiter ./cmd/ratelimiter/main.go \
    && go clean -cache -modcache

# Runtime stage
FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/ratelimiter .
COPY --from=builder /app/internal/storage/pg/migrations ./migrations
COPY --from=builder /app/configs .

RUN apk --no-cache add ca-certificates
EXPOSE 8080
EXPOSE 9000

CMD ["./ratelimiter"]