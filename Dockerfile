# --- Stage 1: build the React frontend ---
FROM node:22-alpine AS frontend-build
WORKDIR /app/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# --- Stage 2: build the Go binary ---
FROM golang:1.26-alpine AS backend-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY config/ ./config/
RUN CGO_ENABLED=0 go build -o /app/bin/server ./cmd/server

# --- Stage 3: minimal runtime image ---
FROM alpine:3.20
# ca-certificates is required for outbound TLS: Yahoo Finance, MongoDB
# Atlas, and Upstash Redis (rediss://) all need valid root certs.
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-build /app/bin/server ./server
COPY --from=frontend-build /app/web/dist ./web/dist

ENV HTTP_ADDR=:8080
EXPOSE 8080
CMD ["./server"]
