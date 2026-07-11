FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api

# ---

FROM alpine:3.23 AS runner

COPY --from=builder /app/bin/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]
