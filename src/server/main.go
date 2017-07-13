package main

import (
	"flag"
	"fmt"
	"gophers"
	"log"
	"net"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
)

var port = flag.Int("port", 12345, "port to listen on")

func main() {
	flag.Parse()

	s := grpc.NewServer()
	srv := &Server{}
	gophers.RegisterGopherAPIServer(s, srv)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("starting gopher server on port: %d", *port)
	log.Fatal(s.Serve(lis))
}

type Server struct{}

func (s *Server) GetGopher(ctx context.Context, g *gophers.GopherRequest) (*gophers.Gopher, error) {
	log.Println("serving up a gopher!")
	return &gophers.Gopher{
		Id:   g.Id,
		Name: "jason",
		Age:  32,
	}, nil
}

func (s *Server) StreamGophers(e *empty.Empty, server gophers.GopherAPI_StreamGophersServer) error {
	gopher := &gophers.Gopher{
		Id:   "some-gopher-id",
		Name: "jason",
		Age:  32,
	}
	for {
		log.Println("serving up a streaming gopher!")
		err := server.Send(gopher)
		if err != nil {
			break
		}
		time.Sleep(time.Second)
	}
	log.Println("done sending gophers")
	return nil
}

func (s *Server) UploadGophers(server gophers.GopherAPI_UploadGophersServer) error {
	for {
		_, err := server.Recv()
		if err != nil {
			break
		}
		log.Println("received a streaming gopher!")
	}
	log.Println("done receiving gophers")
	return nil
}

func (s *Server) GopherChat(server gophers.GopherAPI_GopherChatServer) error {
	go func() {
		for {
			log.Print("sending chat message")
			err := server.Send(&gophers.ChatMessage{
				Content: "hello from server",
			})
			if err != nil {
				return
			}
			time.Sleep(time.Second)
		}
	}()

	for {
		msg, err := server.Recv()
		if err != nil {
			break
		}
		log.Printf("received chat message: %s", msg.Content)
	}
	log.Println("done chatting")
	return nil
}
