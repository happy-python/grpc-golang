package main

import (
	"context"
	"log"

	"grpc-golang/greet/proto"

	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"time"
)

// unary
func doUnary(client greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Stephane",
			LastName:  "Maarek",
		},
	}
	resp, err := client.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling doUnary: %v\n", err)
	}

	log.Println(resp.GetResult())
}

// server streaming
func doServerStreaming(client greetpb.GreetServiceClient) {
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Stephane",
			LastName:  "Maarek",
		},
	}
	stream, err := client.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling doServerStreaming: %v\n", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while doServerStreaming recv: %v\n", err)
		}
		log.Println(resp.GetResult())
	}
}

// client streaming
func doClientStreaming(client greetpb.GreetServiceClient) {
	stream, err := client.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling doClientStreaming: %v\n", err)
	}

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Jack",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Rose",
			},
		},
	}

	for _, req := range requests {
		stream.Send(req)
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while doClientStreaming recv: %v\n", err)
	}

	log.Println(resp.GetResult())
}

// bi directional streaming
func doBiDirectionalStreaming(client greetpb.GreetServiceClient) {
	stream, err := client.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("error while calling doBiDirectionalStreaming: %v\n", err)
	}

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Jack",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Rose",
			},
		},
	}

	waitChan := make(chan struct{})

	go func() {
		for _, req := range requests {
			stream.Send(req)
		}

		stream.CloseSend()
	}()

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitChan)
				break
			}
			if err != nil {
				close(waitChan)
				log.Fatalf("error while doBiDirectionalStreaming recv: %v\n", err)
			}
			log.Println(resp.GetResult())
		}
	}()

	// 阻塞等待
	<-waitChan
}

// unary with deadline
func doUnaryWithDeadline(client greetpb.GreetServiceClient, seconds time.Duration) {
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Stephane",
			LastName:  "Maarek",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), seconds)
	defer cancel()

	resp, err := client.GreetWithDeadline(ctx, req)
	if err != nil {
		errStatus, ok := status.FromError(err)
		if ok {
			if errStatus.Code() == codes.DeadlineExceeded {
				log.Println("Timeout was hit! Deadline was exceeded")
			} else {
				log.Printf("unexpected error: %v\n", errStatus)
			}
		} else {
			log.Fatalf("error while calling doUnaryWithDeadline: %v\n", err)
		}
		return
	}

	log.Println(resp.GetResult())
}

func main() {
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}

	defer conn.Close()

	client := greetpb.NewGreetServiceClient(conn)

	// doUnary(client)
	// doServerStreaming(client)
	// doClientStreaming(client)
	// doBiDirectionalStreaming(client)
	doUnaryWithDeadline(client, 5*time.Second)
	doUnaryWithDeadline(client, 1*time.Second)
}
