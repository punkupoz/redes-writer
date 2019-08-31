FROM golang:1.12-alpine

WORKDIR   /build
COPY    . /build/

RUN apk add --no-cache git
RUN GO111MODULE=on go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -o /app /build/cmd/main.go

FROM alpine:3.9

COPY --from=0 /app /app
RUN apk add --no-cache ca-certificates

EXPOSE 8484

ENTRYPOINT ["/app"]
