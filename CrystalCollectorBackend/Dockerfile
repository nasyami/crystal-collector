# ---- Build Stage ----
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o crystal-collector-backend main.go

# ---- Run Stage ----
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/crystal-collector-backend .
COPY .env .
EXPOSE 8081
ENV PORT=8081
CMD ["/app/crystal-collector-backend"]
