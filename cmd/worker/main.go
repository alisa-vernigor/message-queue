package main

import (
	"log"

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
	if start == finish {
		return 0, []string{start}, nil
	}

	prev := make(map[string]string)
	var queue []string

	queue = append(queue, start)

	foundFinish := false

	for len(queue) > 0 {
		top := queue[0]

		neighbours, err := grabber.Grab(top)
		if err != nil {
			return 0, nil, err
		}

		for _, neighbour := range neighbours {
			if _, ok := prev[neighbour]; !ok {
				prev[neighbour] = top
			}
			if neighbour == finish {
				foundFinish = true
				break
			}
		}

		if foundFinish {
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
	return len(path), reverse(path), nil
}

func main() {
	queues := queue.NewQueues("amqp://guest:guest@localhost:5672/")

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

		len, path, err := bfs(req.StartLink, req.FinishLink)
		if err != nil {
			failOnError(err, "can't find path")
		}

		var resp pb.GetPathResponse
		resp.Path = path
		resp.PathLength = int64(len)
		resp.Uuid = req.Uuid

		body, err := proto.Marshal(&resp)

		if err != nil {
			failOnError(err, "can't find path")
		}

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
	}
}
