FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY proto/ proto/
COPY gateway-service/ gateway-service/

WORKDIR /build/gateway-service

ENV GOTOOLCHAIN=auto
RUN go mod download
RUN CGO_ENABLED=0 go build -o /bin/gateway-service ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/gateway-service /bin/gateway-service

EXPOSE 8000

CMD ["/bin/gateway-service"]
