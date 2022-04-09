package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	pb "github.com/alisa-vernigor/message-queue/proto/pathfinder"
	"google.golang.org/grpc"
)

func main() {
	var conn *grpc.ClientConn
	var err error
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Input addr:")
		addr, _ := reader.ReadString('\n')
		addr = strings.Replace(addr, "\n", "", -1)
		conn, err = grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			fmt.Println("Wrong adress, try again")
		}
		break
	}
	defer conn.Close()

	fmt.Println("Enter start URL:")
	url1, _ := reader.ReadString('\n')
	url1 = strings.Replace(url1, "\n", "", -1)
	fmt.Println("Enter finish URL:")
	url2, _ := reader.ReadString('\n')
	url2 = strings.Replace(url2, "\n", "", -1)

	c := pb.NewPathFinderClient(conn)

	resp, err := c.GetPath(context.Background(), &pb.GetPathRequest{StartLink: url1, FinishLink: url2})
	fmt.Printf("%s, %d hops\n", strings.Join(resp.Path, " => "), resp.GetPathLength())
}
