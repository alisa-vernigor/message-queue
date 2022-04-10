FROM golang:latest

RUN apt-get update

WORKDIR /go/src/

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

EXPOSE 9090

CMD ["/bin/bash",  "-c", "while ! curl -s rabbitmq:15672 > /dev/null; do echo waiting for rabbitmq; sleep 3; done; go run ./cmd/pathfinder/main.go"]