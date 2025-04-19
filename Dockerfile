# builder
FROM golang:1.24 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# runtime
FROM alpine:3.21 AS runtime

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/server .

USER appuser

CMD ["/app/server"]
