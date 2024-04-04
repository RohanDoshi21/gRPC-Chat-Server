## gRPC Chat Server
This is a simple gRPC chat server implemented in Go. It uses the gRPC framework for handling real-time communication between clients. It is one-to-one chat server, meaning that a client can send messages to another client by specifying the receiver's username.

### Features
`Real-time communication`: The server can handle multiple clients and facilitate real-time communication between them.

`Message confirmation`: The server sends a confirmation message back to the sender after successfully forwarding the message to the receiver.

`Connection handling`: The server handles client disconnections gracefully by removing the client from the active clients list.


### How it works
The server maintains a map of active clients. Each client is identified by a unique username. When a client sends a message, the server forwards the message to the intended receiver and sends a confirmation back to the sender. If the receiver is not found in the active clients list, the server simply ignores the message.

### Usage
To run the server, simply execute the main.go file:
```bash
go run main.go
```

The server will start and listen for incoming connections on `port 8080`.