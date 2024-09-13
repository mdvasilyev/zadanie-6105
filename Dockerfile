FROM golang:latest

RUN apt-get update

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.12.2/migrate.linux-amd64.tar.gz | tar xvz

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o app ./cmd/zadanie-6105

COPY ./internal/app/database/migration /app/migrations

RUN migrate -path /app/migrations -database "$POSTGRES_CONN" -verbose up

EXPOSE 8080

CMD ["./app"]
