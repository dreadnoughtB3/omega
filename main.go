package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:54321", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Tetragon: %v", err)
	}
	defer conn.Close()

	client := tetragon.NewFineGuidanceSensorsClient(conn)
	req := &tetragon.GetEventsRequest{}

	stream, err := client.GetEvents(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}

	fmt.Println("Listening for Tetragon events...")

	// イベントループ
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Stream error: %v", err)
		}

		// イベントの種類に応じて処理
		switch event := res.Event.(type) {
		case *tetragon.GetEventsResponse_ProcessExec:
			proc := event.ProcessExec.Process
			fmt.Printf("🚀 EXEC: %s (PID: %d) in Pod: %s\n", proc.Binary, proc.Pid, proc.Pod.Name)

		case *tetragon.GetEventsResponse_ProcessExit:
			proc := event.ProcessExit.Process
			fmt.Printf("💥 EXIT: %s (PID: %d) Status: %d\n", proc.Binary, proc.Pid, event.ProcessExit.Status)
		}
	}
}
