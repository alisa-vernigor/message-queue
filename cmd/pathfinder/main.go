package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/alisa-vernigor/message-queue/proto/pathfinder"
	"github.com/streadway/amqp"
)

type pathFinder struct{}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (p *pathFinder) GetPath(c context.Context, req *pb.GetPathRequest) (*pb.GetPathResponse, error) {
	return nil, nil
}

func get_adress() string {
	conn, error := net.Dial("udp", "8.8.8.8:80")
	if error != nil {
		fmt.Println(error)

	}

	defer conn.Close()
	ipAddress := conn.LocalAddr().(*net.UDPAddr)
	return ipAddress.IP.String()
}

func main() {
	var lis net.Listener
	var err error
	var addr string

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for {
		addr = "9090"

		addr = ":" + addr
		lis, err = net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		break
	}
	fmt.Println("Listening on:", get_adress()+addr)

	s := grpc.NewServer()

	pb.RegisterChatRoomServer(s, &pathFinder{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
