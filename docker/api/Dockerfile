# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY services/api/go.mod services/api/go.sum ./
RUN go mod download

COPY services/api/ .
RUN CGO_ENABLED=0 go build -o api .

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/api ./api
COPY --from=builder /build/migrations ./migrations

ENV PORT=8080
ENV MIGRATIONS_PATH=/app/migrations

EXPOSE 8080

CMD ["./api"]
