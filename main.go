package main

import (
	"io"
	"log"
	pb "main/protocol"
	"net"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.GrpcServerServiceServer
	clients map[string]pb.GrpcServerService_SendMessageServer
	mu      sync.Mutex
}

func main() {
	server := &Server{
		clients: make(map[string]pb.GrpcServerService_SendMessageServer),
	}

	grpcServer := grpc.NewServer()

	pb.RegisterGrpcServerServiceServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error creating server", err)
	}

	log.Printf("gRPC server listening on %s", ":8080")
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Error serving gRPC server", err)
	}
}

type Payload struct {
	Username string
}

func (server *Server) SendMessage(stream pb.GrpcServerService_SendMessageServer) error {
	payload := &Payload{}

	for {
		message, err := stream.Recv()

		if err == io.EOF {
			// The client has closed the connection.
			break
		}
		payload.Username = message.Username

		if server.clients[payload.Username] == nil {
			server.mu.Lock()
			server.clients[payload.Username] = stream
			server.mu.Unlock()
		}

		if err != nil {
			return status.Errorf(codes.Internal, "Error receiving message: %v", err)
		}

		// Find the receiver by username.
		server.mu.Lock()
		receiver, ok := server.clients[message.Reciever]
		if !ok {
			// If the receiver or sender is not found, send an error message back to the sender.
			// Avoid Deadlock of the server
			server.mu.Unlock()
			continue
		}

		sender, ok := server.clients[payload.Username]
		server.mu.Unlock()

		if !ok {
			// If the receiver or sender is not found, send an error message back to the sender.
			continue
		}

		messageUUID := uuid.New().String()

		// Forward the message to the receiver.
		err = receiver.Send(&pb.Message{
			Sender:   payload.Username,
			Receiver: message.Reciever,
			Message:  message.Message,
			Id:       messageUUID,
		})
		if err != nil {
			log.Printf("Error sending message to %s: %v", message.Reciever, err)
			continue
		}

		// Send the same message back to the sender as a confirmation.
		err = sender.Send(&pb.Message{
			Sender:   payload.Username,
			Receiver: message.Reciever,
			Message:  message.Message,
			Id:       messageUUID,
		})
		if err != nil {
			log.Printf("Error sending confirmation message to %s: %v", payload.Username, err)
			continue
		}
	}

	// Remove the sender from the clients map when the client disconnects.
	server.mu.Lock()
	delete(server.clients, payload.Username)
	server.mu.Unlock()
	return nil
}
