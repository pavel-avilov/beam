// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	pb "beam.apache.org/playground/backend/internal/api"
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	bufSize    = 1024 * 1024
	codeString = "class HelloWorld {\n    public static void main(String[] args) {\n        System.out.println(\"Hello World!\");\n    }\n}"
)

var lis *bufconn.Listener

func TestMain(m *testing.M) {
	server := setup()
	defer teardown(server)
	m.Run()
}

func setup() *grpc.Server {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterPlaygroundServiceServer(s, &playgroundController{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return s
}

func teardown(server *grpc.Server) {
	server.Stop()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}
func TestPlaygroundController_RunCode(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewPlaygroundServiceClient(conn)
	code := pb.RunCodeRequest{
		Code: "test",
		Sdk:  pb.Sdk_SDK_JAVA,
	}
	pipelineMeta, err := client.RunCode(ctx, &code)
	if err != nil {
		t.Fatalf("runCode failed: %v", err)
	}
	log.Printf("Response: %+v", pipelineMeta)
}

func TestPlaygroundController_CheckStatus(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewPlaygroundServiceClient(conn)
	pipelineMeta := pb.CheckStatusRequest{
		PipelineUuid: uuid.NewString(),
	}
	status, err := client.CheckStatus(ctx, &pipelineMeta)
	if err != nil {
		t.Fatalf("runCode failed: %v", err)
	}
	log.Printf("Response: %+v", status)
}

func TestPlaygroundController_GetCompileOutput(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := pb.NewPlaygroundServiceClient(conn)
	pipelineMeta := pb.GetCompileOutputRequest{
		PipelineUuid: uuid.NewString(),
	}
	compileOutput, err := client.GetCompileOutput(ctx, &pipelineMeta)
	if err != nil {
		t.Fatalf("runCode failed: %v", err)
	}
	log.Printf("Response: %+v", compileOutput)
}

func Test_playgroundController_GetRunOutput(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewPlaygroundServiceClient(conn)
	pipelineId := executeCode(t, client, ctx)

	wrongUuid := uuid.NewString()
	for pipelineId == wrongUuid {
		wrongUuid = uuid.NewString()
	}

	type args struct {
		ctx  context.Context
		info *pb.GetRunOutputRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.GetRunOutputResponse
		wantErr bool
	}{
		{
			name: "GetRunOutput with right pipelineId",
			args: args{
				ctx:  ctx,
				info: &pb.GetRunOutputRequest{PipelineUuid: pipelineId},
			},
			want:    &pb.GetRunOutputResponse{Output: "Hello World!\n"},
			wantErr: false,
		},
		{
			name: "GetRunOutput with wrong pipelineId",
			args: args{
				ctx:  ctx,
				info: &pb.GetRunOutputRequest{PipelineUuid: wrongUuid},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.GetRunOutput(tt.args.ctx, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRunOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !strings.EqualFold(got.Output, tt.want.Output) {
					t.Errorf("GetRunOutput() got = %v, want %v", got.Output, tt.want.Output)
				}
				if !reflect.DeepEqual(got.CompilationStatus, tt.want.CompilationStatus) {
					t.Errorf("GetRunOutput() got = %v, want %v", got.CompilationStatus, tt.want.CompilationStatus)
				}
			}
		})
	}
}

// executeCode sends RunCodeRequest and waits until the end of the execution.
// Returns pipelineId or fails the test where it is used if error occurs
func executeCode(t *testing.T, client pb.PlaygroundServiceClient, ctx context.Context) string {
	code := pb.RunCodeRequest{
		Code: codeString,
		Sdk:  pb.Sdk_SDK_JAVA,
	}
	pipelineMeta, err := client.RunCode(ctx, &code)
	if err != nil {
		t.Fatalf("runCode failed: %v", err)
	}

	pipelineId := pipelineMeta.PipelineUuid
	runOutputRequest := pb.GetRunOutputRequest{PipelineUuid: pipelineId}
	timeToWait := time.Second
	for timeToWait < time.Second*30 {
		time.Sleep(timeToWait)
		timeToWait *= 2
		_, err := client.GetRunOutput(ctx, &runOutputRequest)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("runCode took too much time: %v", err)
	}
	return pipelineId
}
