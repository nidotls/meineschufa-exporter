FROM golang:1.22 as builder

ENV GOOS linux
ENV CGO_ENABLED 0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o meineschufa-exporter

FROM alpine:3.14 as production

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/meineschufa-exporter /bin/meineschufa-exporter

CMD meineschufa-exporter