FROM golang:1.16.3-alpine3.13 AS builder

WORKDIR /app

COPY . .

RUN go get -d -v ./...
RUN go build -o main .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]