FROM golang:latest

RUN apt-get update && apt-get install -y curl gnupg lsb-release ca-certificates

RUN curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -

RUN echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/migrate.list

RUN apt-get update && apt-get install -y migrate

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o app ./cmd/zadanie-6105

COPY ./internal/app/database/migration /app/migrations

RUN migrate -path /app/migrations -database "$POSTGRES_CONN" -verbose up

EXPOSE 8080

CMD ["./app"]
