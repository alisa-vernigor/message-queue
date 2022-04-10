package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/alisa-vernigor/message-queue/internal/queue"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"

	pb "github.com/alisa-vernigor/message-queue/proto/pathfinder"
)

type pathFinder struct {
	pb.UnimplementedPathFinderServer
	channels map[string]chan *pb.GetPathResponse
	queues   queue.Queues
	mu       sync.RWMutex
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func (p *pathFinder) GetPath(c context.Context, req *pb.GetPathRequest) (*pb.GetPathResponse, error) {
	log.Println("Got request:", req.StartLink, '\n', req.FinishLink)

	id := uuid.New().String()
	log.Println("uuid:", id)

	ch := make(chan *pb.GetPathResponse)

	p.mu.Lock()
	p.channels[id] = ch
	p.mu.Unlock()

	req.Uuid = id
	body, err := proto.Marshal(req)

	if err != nil {
		return nil, err
	}

	log.Println("Ready to publish to queue")
	p.queues.Ch.Publish(
		"",     // exchange
		"task", // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         body,
		},
	)

	log.Println("Published, waiting for response")

	resp := <-ch

	log.Println("Got response, returning!")

	return resp, nil
}

func newPathFinder() *pathFinder {
	var path pathFinder

	path.queues = queue.NewQueues("amqp://guest:guest@rabbitmq:5672/")
	path.channels = make(map[string]chan *pb.GetPathResponse)
	path.mu = sync.RWMutex{}

	go path.respGetter()

	return &path
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

func (p *pathFinder) respGetter() {

	msgs, err := p.queues.Ch.Consume(
		"result", // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	failOnError(err, "Failed to register a consumer")
	for d := range msgs {

		var resp pb.GetPathResponse
		proto.Unmarshal(d.Body, &resp)

		log.Println("Got ans:", resp.Uuid)

		p.mu.Lock()
		log.Println("Write to channel")
		p.channels[resp.GetUuid()] <- &resp
		log.Println("Wrote to channel")
		p.mu.Unlock()
	}
}

func main() {
	var lis net.Listener
	var err error
	var addr string

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

	pb.RegisterPathFinderServer(s, newPathFinder())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
