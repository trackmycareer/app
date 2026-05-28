# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /build/server ./cmd/server

# Stage 2: Minimal production image
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
RUN adduser -D -u 1000 appuser

COPY --from=builder /build/server /usr/local/bin/server

USER appuser
EXPOSE 8080
ENTRYPOINT ["server"]
