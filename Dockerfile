FROM golang:1.18 AS builder
WORKDIR /go/src/build
COPY . .
RUN go build -o pod-chaos-monkey ./cli/cmd/main.go

FROM alpine:3.16.2

RUN set -x \
   && addgroup cm \
   && adduser -D -G cm cm \
   && mkdir -p /app \
   && chown cm: /app \
   && apk --no-cache upgrade \
   && apk --no-cache add ca-certificates tzdata

WORKDIR /app
USER cm
COPY --chown=cm --from=builder /go/src/build/pod-chaos-monkey /app
ENTRYPOINT [ "/app/pod-chaos-monkey" ]
