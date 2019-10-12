package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os/signal"

	"google.golang.org/grpc"

	"grpc-golang/blog/proto"

	"os"
)

var collection *mongo.Collection

type service struct {
}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Title    string             `bson:"title"`
	Content  string             `bson:"content"`
}

// 工具函数
func dataToBlogPb(data blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func (s *service) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	blog := req.GetBlog()
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}
	insertResult, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	id, ok := insertResult.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("convert err"))
	}

	data.ID = id

	res := &blogpb.CreateBlogResponse{
		Blog: dataToBlogPb(data),
	}

	return res, nil
}

func (s *service) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	id := req.BlogId
	var data blogItem

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}

	err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&data)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("not found err: %v", err))
	}

	res := &blogpb.ReadBlogResponse{
		Blog: dataToBlogPb(data),
	}

	return res, nil
}

func (s *service) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	blog := req.GetBlog()
	id := blog.GetId()

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}

	filter := bson.M{"_id": oid}

	var data blogItem
	err = collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("not found err: %v", err))
	}

	// UpdateOne
	//update := bson.D{{"$set", bson.D{{"author_id", blog.GetAuthorId()}, {"title", blog.GetTitle()}, {"content", blog.GetContent()}}}}
	//updateResult, err := collection.UpdateOne(context.Background(), filter, update)

	// ReplaceOne
	data.AuthorID = blog.GetAuthorId()
	data.Title = blog.GetTitle()
	data.Content = blog.GetContent()
	updateResult, err := collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal err: %v", err))
	}

	fmt.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)

	res := &blogpb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}

	return res, nil
}

func (s *service) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	id := req.BlogId

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}

	filter := bson.M{"_id": oid}

	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("delete err: %v", err))
	}

	fmt.Printf("Deleted %v documents \n", deleteResult.DeletedCount)
	if deleteResult.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("not found"))
	}

	res := &blogpb.DeleteBlogResponse{
		BlogId: id,
	}

	return res, nil
}

func (s *service) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	// Pass these options to the Find method
	findOptions := options.Find()
	cur, err := collection.Find(context.Background(), bson.D{{}}, findOptions)
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("list find err: %v", err))
	}
	// Close the cursor once finished
	defer cur.Close(context.TODO())

	var results []blogItem

	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var data blogItem
		err := cur.Decode(&data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Error while decoding data from MongoDB: %v", err))
		}

		stream.Send(&blogpb.ListBlogResponse{
			Blog: dataToBlogPb(data),
		})

		results = append(results, data)
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("cur err: %v", err))
	}

	fmt.Printf("Found multiple documents: %+v\n", results)
	return nil
}

func main() {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection = client.Database("blogdb").Collection("blog")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	defer listener.Close()

	log.Println("listen at :50051")

	opt := []grpc.ServerOption{}
	server := grpc.NewServer(opt...)
	blogpb.RegisterBlogServiceServer(server, &service{})

	go func() {
		log.Println("blog server started")
		if err = server.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v\n", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	log.Println("blog server stopped")
}
