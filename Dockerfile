FROM golang:1.22-alpine AS builder

WORKDIR /oengus-bot
COPY go.mod go.sum ./
#COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /oengus-bot
COPY --from=builder /oengus-bot/main ./main
RUN chmod +x main

CMD ["./main"]
