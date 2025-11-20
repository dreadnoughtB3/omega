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
	// 1. Tetragon gRPCサーバーへの接続
	conn, err := grpc.NewClient("localhost:54321", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Tetragon: %v", err)
	}
	defer conn.Close()

	client := tetragon.NewFineGuidanceSensorsClient(conn)

	// 2. イベント監視リクエスト
	req := &tetragon.GetEventsRequest{}

	stream, err := client.GetEvents(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to open stream: %v", err)
	}

	fmt.Println("Listening for Tetragon events...")

	// 3. イベントループ
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
			if event.ProcessExec == nil || event.ProcessExec.Process == nil {
				continue
			}
			proc := event.ProcessExec.Process
			
			// Pod情報のNilチェック
			podName := "Host Process"
			if proc.Pod != nil {
				podName = "Pod: " + proc.Pod.Name
			}

			fmt.Printf("🚀 EXEC: %s (PID: %d) [%s]\n", proc.Binary, proc.Pid, podName)
		
		case *tetragon.GetEventsResponse_ProcessExit:
			if event.ProcessExit == nil || event.ProcessExit.Process == nil {
				continue
			}
			proc := event.ProcessExit.Process

			// Pod情報のNilチェック
			podName := "Host Process"
			if proc.Pod != nil {
				podName = "Pod: " + proc.Pod.Name
			}

			// 異常終了（ステータス0以外）を目立たせる
			status := event.ProcessExit.Status
			if status != 0 {
				fmt.Printf("💥 EXIT (ERROR): %s (PID: %d) Status: %d [%s]\n", proc.Binary, proc.Pid, status, podName)
			} else {
				fmt.Printf("👋 EXIT (OK): %s (PID: %d) [%s]\n", proc.Binary, proc.Pid, podName)
			}
		}
	}
}