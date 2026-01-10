# Heroku-friendly container build for click-service
FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin app
WORKDIR /app
COPY --from=build /out/server ./server
USER app

ENV PORT=8082
EXPOSE 8082

CMD ["./server"]
