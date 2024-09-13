FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o app ./cmd/zadanie-6105

EXPOSE 8080

CMD ["./app"]
