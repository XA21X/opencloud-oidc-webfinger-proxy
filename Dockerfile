FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod init opencloud-oidc-webfinger-proxy || true
RUN go mod tidy
RUN go build -o opencloud-oidc-webfinger-proxy main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/opencloud-oidc-webfinger-proxy .

EXPOSE 9210
ENV PORT=9210
ENV REDIRECT_TEMPLATE=""

CMD ["./opencloud-oidc-webfinger-proxy"]