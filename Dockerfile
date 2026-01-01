FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 go build -o zip-agent .

FROM alpine:latest

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY --from=builder /app/zip-agent .

EXPOSE 8080

CMD ["./zip-agent"]
