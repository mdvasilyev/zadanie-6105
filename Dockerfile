FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod tidy

RUN  #curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add - \
#     echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/migrate.list \
#     apt-get update \
#     apt-get install -y migrate

RUN go build -o app ./cmd/zadanie-6105

COPY ./internal/app/database/migration /app/migrations

EXPOSE 8080

RUN #migrate -path /app/migrations -database "$POSTGRES_CONN" -verbose up

CMD ["./app"]
