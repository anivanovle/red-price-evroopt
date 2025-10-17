
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o app cmd/main.go
FROM alpine:latest
RUN apk add --no-cache chromium nss ca-certificates 
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/migrations ./migrations
ENV CHROME_BIN=/usr/bin/chromium-browser
CMD ["./app"]
