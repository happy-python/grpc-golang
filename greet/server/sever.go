package main

import (
	"context"
	"log"
	"net"

	"grpc-golang/greet/proto"

	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"io"

	"time"
)

type service struct{}

// unary
func (s *service) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Println("starting Greet")

	firstName := req.GetGreeting().GetFirstName()

	result := "Hello " + firstName

	return &greetpb.GreetResponse{
		Result: result,
	}, nil
}

// server streaming
func (s *service) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	log.Println("starting GreetManyTimes")

	firstName := req.GetGreeting().GetFirstName()
	for index := 0; index < 10; index++ {
		result := "Hello " + firstName + " number " + strconv.Itoa(index)
		stream.Send(&greetpb.GreetManyTimesResponse{
			Result: result,
		})
	}

	return nil
}

// client streaming
func (s *service) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	log.Println("starting LongGreet")

	result := ""

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("failed to LongGreet: %v\n", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result += "Hello " + firstName + "! "
	}
}

// bi directional streaming
func (s *service) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	log.Println("starting GreetEveryone")

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("failed to GreetEveryone: %v\n", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "! "
		stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
	}
}

// unary with deadline
func (s *service) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	log.Println("starting GreetWithDeadline")

	for index := 0; index < 3; index++ {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("The client canceled the request")
			return nil, status.Error(codes.Canceled, "Client cancelled, abandoning.")
		}

		time.Sleep(1 * time.Second)
	}

	firstName := req.GetGreeting().GetFirstName()

	result := "Hello " + firstName

	return &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	defer listener.Close()

	log.Println("listen at :50051")

	server := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(server, &service{})

	if err = server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
