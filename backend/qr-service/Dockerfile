# Heroku-friendly container build for qr-service
FROM golang:1.24-alpine AS build

WORKDIR /app

# Faster + reproducible deps
COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin app
WORKDIR /app
COPY --from=build /out/server ./server
USER app

# Heroku sets PORT; the server reads it.
ENV PORT=8080
EXPOSE 8080

CMD ["./server"]
