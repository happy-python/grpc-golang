package main

import (
	"context"
	"google.golang.org/grpc"
	"grpc-golang/blog/proto"
	"io"
	"log"
)

var client blogpb.BlogServiceClient

func CreateBlog() {
	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "jack",
			Title:    "My First Blog",
			Content:  "content of first blog",
		},
	}

	res, err := client.CreateBlog(context.TODO(), req)
	if err != nil {
		log.Fatalf("create blog err: %v\n", err)
	}
	log.Printf("create blog: %v", res)
}

func ReadBlog() {
	req := &blogpb.ReadBlogRequest{
		BlogId: "5d9ffd670071746caaf4a25f",
	}

	res, err := client.ReadBlog(context.TODO(), req)
	if err != nil {
		log.Fatalf("read blog err: %v\n", err)
	}
	log.Printf("read blog: %v", res)
}

func UpdateBlog() {
	req := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       "5d9ffd670071746caaf4a25f",
			AuthorId: "jack",
			Title:    "My First Blog(edited)",
			Content:  "content of first updated blog(edited)",
		},
	}

	res, err := client.UpdateBlog(context.TODO(), req)
	if err != nil {
		log.Fatalf("update blog err: %v\n", err)
	}
	log.Printf("update blog: %v", res)
}

func DeleteBlog()  {
	req := &blogpb.DeleteBlogRequest{
		BlogId: "5d9ffd670071746caaf4a25f",
	}

	res, err := client.DeleteBlog(context.TODO(), req)
	if err != nil {
		log.Fatalf("delete blog err: %v\n", err)
	}
	log.Printf("delete blog: %v", res)
}

func ListBlog()  {
	req := &blogpb.ListBlogRequest{}

	stream, err := client.ListBlog(context.TODO(), req)
	if err != nil {
		log.Fatalf("list blog err: %v\n", err)
	}

	waitChan := make(chan struct{})

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitChan)
				break
			}
			if err != nil {
				close(waitChan)
				log.Fatalf("list blog recv err: %v\n", err)
			}
			log.Printf("list blog: %v", res.GetBlog())
		}
	}()

	<-waitChan
}

func main() {
	opt := grpc.WithInsecure()
	conn, err := grpc.Dial(":50051", opt)
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}

	defer conn.Close()

	client = blogpb.NewBlogServiceClient(conn)

	//CreateBlog()
	//ReadBlog()
	//UpdateBlog()
	//DeleteBlog()
	ListBlog()
}
