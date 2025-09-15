
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-s -w" -o altavin-bot main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/altavin-bot ./
EXPOSE 8090
CMD ["./altavin-bot"]
