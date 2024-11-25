FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CONFIG_DIR=/app/config

WORKDIR /app/cmd/api

RUN go build -o /myapp

EXPOSE 8080

CMD ["/myapp"]