FROM golang:latest

RUN apt-get update

RUN wget http://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-arm64.deb

RUN dpkg -i migrate.linux-arm64.deb

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o app ./cmd/zadanie-6105

COPY ./internal/app/database/migration /app/migrations

RUN migrate -path /app/migrations -database "$POSTGRES_CONN" -verbose up

EXPOSE 8080

CMD ["./app"]
