package main

import (
	"log"
	"os"

	"github.com/alisa-vernigor/message-queue/internal/grabber"
	"github.com/alisa-vernigor/message-queue/internal/queue"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"

	pb "github.com/alisa-vernigor/message-queue/proto/pathfinder"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func reverse(arr []string) []string {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func bfs(start string, finish string) (int, []string, error) {
	log.Println("Starting bfs!")
	if start == finish {
		return 0, []string{start}, nil
	}

	prev := make(map[string]string)
	var queue []string

	queue = append(queue, start)

	foundFinish := false

	for len(queue) > 0 {
		top := queue[0]
		queue = queue[1:]

		neighbours, err := grabber.Grab(top)
		//log.Println("Got links:", neighbours)

		if err != nil {
			return 0, nil, err
		}

		for _, neighbour := range neighbours {
			if _, ok := prev[neighbour]; !ok {
				prev[neighbour] = top
				queue = append(queue, neighbour)
			}
			if neighbour == finish {
				log.Println("Found answer, breaking!")
				foundFinish = true
				break
			}
		}

		if foundFinish {
			log.Println("Found answer, breaking!")
			break
		}
	}

	var path []string
	cur := finish
	for cur != start {
		path = append(path, cur)
		cur = prev[cur]
	}
	path = append(path, start)
	log.Println("Returning from bfs!")
	return len(path), reverse(path), nil
}

func main() {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	log.SetOutput(f)
	queues := queue.NewQueues("amqp://guest:guest@rabbitmq:5672/")

	msgs, err := queues.Ch.Consume(
		"task", // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")
	for d := range msgs {
		var req pb.GetPathRequest

		proto.Unmarshal(d.Body, &req)

		log.Println("Got request:", req.StartLink, '\n', req.FinishLink)
		log.Println("Ready for bfs!")

		len, path, err := bfs(req.StartLink, req.FinishLink)
		if err != nil {
			failOnError(err, "can't find path")
		}
		log.Println("Found answer!")

		var resp pb.GetPathResponse
		resp.Path = path
		resp.PathLength = int64(len)
		resp.Uuid = req.Uuid

		body, err := proto.Marshal(&resp)

		if err != nil {
			failOnError(err, "can't find path")
		}

		log.Println("Ready to push to queue!")

		queues.Ch.Publish(
			"",       // exchange
			"result", // routing key
			false,    // mandatory
			false,    // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "text/plain",
				Body:         body,
			},
		)
		log.Println("Oushed to queue")
	}
}
