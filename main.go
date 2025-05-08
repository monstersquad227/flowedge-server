package main

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/google/uuid"
	pb "github.com/monstersquad227/flowedge-proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type server struct {
	pb.UnimplementedFlowEdgeServer
	streams         sync.Map
	pendingResponse sync.Map
}

func (s *server) Communicate(stream pb.FlowEdge_CommunicateServer) error {
	var agentID string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Agent %s disconnected", agentID)
			if agentID != "" {
				s.streams.Delete(agentID)
			}
			return nil
		}
		if err != nil {
			log.Printf("Recv error from agent %s: %v", agentID, err)
			if agentID != "" {
				s.streams.Delete(agentID)
			}
			return err
		}

		switch msg.Type {
		case pb.MessageType_REGISTER:
			log.Printf("Register: %+v", msg.GetRegister())
			agentID = msg.GetRegister().AgentId
			s.streams.Store(agentID, stream)
			log.Printf("Agent %s registered", agentID)

		case pb.MessageType_HEARTBEAT:
			log.Printf("Heartbeat from %s", msg.GetHeartbeat().AgentId)

		case pb.MessageType_EXECUTE_RESPONSE:
			r := msg.GetExecuteResponse()
			log.Printf("Execution result: %s, output: %s, error: %s", r.CommandId, r.Output, r.Error)

			if chVal, ok := s.pendingResponse.Load(r.CommandId); ok {
				ch := chVal.(chan *pb.ExecuteResponse)
				ch <- r
				s.pendingResponse.Delete(r.CommandId)
			}
		}
	}
}

func main() {
	Server := &server{}
	// 启动 HTTP 服务
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		agentID := r.URL.Query().Get("agent_id")
		command := r.URL.Query().Get("command")
		containerID := r.URL.Query().Get("container_id")
		image := r.URL.Query().Get("image")
		//if agentID == "" || command == "" {
		//	http.Error(w, "agent_id and command required", http.StatusBadRequest)
		//	return
		//}
		streamVal, ok := Server.streams.Load(agentID)
		if !ok {
			http.Error(w, "agent not connected", http.StatusNotFound)
			return
		}
		commandID := uuid.New().String()

		respCh := make(chan *pb.ExecuteResponse, 1)
		Server.pendingResponse.Store(commandID, respCh)

		stream := streamVal.(pb.FlowEdge_CommunicateServer)
		err := stream.Send(&pb.StreamMessage{
			Type: pb.MessageType_EXECUTE_REQUEST,
			Body: &pb.StreamMessage_ExecuteRequest{
				ExecuteRequest: &pb.ExecuteRequest{
					CommandId:   commandID,
					Command:     command,
					Image:       image,
					ContainerId: containerID,
				},
			},
		})
		if err != nil {
			http.Error(w, "failed to send command: "+err.Error(), http.StatusInternalServerError)
			return
		}

		select {
		case resp := <-respCh:
			w.Write([]byte("Command executed\nOutput: " + resp.Output + "\nError: " + resp.Error))
		case <-time.After(20 * time.Second):

			Server.pendingResponse.Delete(commandID)
			http.Error(w, "timeout waiting for agent response", http.StatusGatewayTimeout)
		}
		//w.Write([]byte("Command sent to agent " + agentID))
	})

	go func() {
		log.Println("HTTP server listening on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// 加载服务端证书
	cert, err := tls.LoadX509KeyPair("./certs/server.crt", "./certs/server.key")
	if err != nil {
		log.Fatalf("Failed to load server cert: %v", err)
	}
	// 加载客户端 CA 证书
	caCert, err := ioutil.ReadFile("./certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)
	// 创建 TLS 配置
	creeds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	})

	s := grpc.NewServer(grpc.Creds(creeds))

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	pb.RegisterFlowEdgeServer(s, Server)
	log.Println("Server listening on :50051")
	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
