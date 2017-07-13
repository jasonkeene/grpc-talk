package main

import (
	"context"
	"flag"
	"gophers"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	"google.golang.org/grpc"
)

var target = flag.String("target", "localhost:12345", "server host and port")

func main() {
	flag.Parse()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	var c gophers.GopherAPIClient

	// establish client
	{
		conn, err := grpc.Dial(*target, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		c = gophers.NewGopherAPIClient(conn)
	}

	// request/response
	{
		gopher, err := c.GetGopher(context.TODO(), &gophers.GopherRequest{
			Id: "some-gopher-id",
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("received gopher: %#v", gopher)
	}
	<-sigint

	// streaming server->client
	{
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := c.StreamGophers(ctx, &empty.Empty{})
		if err != nil {
			log.Fatal(err)
		}
	stream_exit:
		for {
			select {
			case <-sigint:
				break stream_exit
			default:
			}
			gopher, err := stream.Recv()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("received gopher: %#v", gopher)
		}
		cancel()
	}
	<-sigint

	// streaming client->server
	{
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := c.UploadGophers(ctx)
		if err != nil {
			log.Fatal(err)
		}
	upload_exit:
		for {
			select {
			case <-sigint:
				break upload_exit
			default:
			}
			log.Print("sending gopher")
			err := stream.Send(&gophers.Gopher{})
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Second)
		}
		cancel()
	}
	<-sigint

	// bi-directional streaming
	{
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := c.GopherChat(ctx)
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			for {
				log.Print("sending chat message")
				err := stream.Send(&gophers.ChatMessage{
					Content: "hello from client",
				})
				if err != nil {
					return
				}
				time.Sleep(time.Second)
			}
		}()
		go func() {
			for {
				msg, err := stream.Recv()
				if err != nil {
					return
				}
				log.Printf("received chat message: %s", msg.Content)
			}
		}()
		<-sigint
		cancel()
	}
}
