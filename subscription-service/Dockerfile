FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY proto/ proto/
COPY subscription-service/ subscription-service/

WORKDIR /build/subscription-service

ENV GOTOOLCHAIN=auto
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/subscription-service ./cmd/server
RUN CGO_ENABLED=0 go build -o /bin/subscription-migrate ./cmd/migration

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/subscription-service /bin/subscription-service
COPY --from=builder /bin/subscription-migrate /bin/subscription-migrate
COPY subscription-service/docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

EXPOSE 50052

ENTRYPOINT ["/docker-entrypoint.sh"]
