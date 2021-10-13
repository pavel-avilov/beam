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
	"beam.apache.org/playground/backend/internal/cache"
	"beam.apache.org/playground/backend/internal/errors"
	"beam.apache.org/playground/backend/internal/executors"
	"beam.apache.org/playground/backend/internal/fs_tool"
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc/grpclog"
	"os"
	"time"
)

type playgroundController struct {
	pb.UnimplementedPlaygroundServiceServer
}

var cacheService = cache.NewCache(os.Getenv("cache"))

//RunCode is running code from requests using a particular SDK
func (controller *playgroundController) RunCode(ctx context.Context, info *pb.RunCodeRequest) (*pb.RunCodeResponse, error) {
	pipelineId := uuid.New()
	expTime, err := time.ParseDuration(os.Getenv("expiration"))
	if err != nil {
		expTime = time.Hour
	}

	// create file system service
	lc, err := fs_tool.NewLifeCycle(info.Sdk, pipelineId)
	if err != nil {
		grpclog.Error("RunCode: NewLifeCycle: " + err.Error())
		return nil, errors.InternalError("Run code", "Error during creating file system service: "+err.Error())
	}

	// create folders
	err = lc.CreateFolders()
	if err != nil {
		grpclog.Error("RunCode: CreateFolders: " + err.Error())
		return nil, errors.InternalError("Run code", "Error during preparing folders: "+err.Error())
	}

	// create file with code
	_, err = lc.CreateExecutableFile(info.Code)
	if err != nil {
		grpclog.Error("RunCode: CreateExecutableFile: " + err.Error())
		return nil, errors.InternalError("Run code", "Error during creating file with code: "+err.Error())
	}

	// create executor
	exec, err := executors.NewExecutor(info.Sdk, lc)
	if err != nil {
		grpclog.Error("RunCode: NewExecutor: " + err.Error())
		return nil, errors.InternalError("Run code", "Error during creating executor: "+err.Error())
	}

	err = cacheService.SetValue(pipelineId, cache.SubKey_Status, pb.Status_STATUS_EXECUTING)
	if err != nil {
		grpclog.Error("RunCode: cache.SetValue: " + err.Error())
		return nil, errors.InternalError("Run code", "Error during set value to cache: "+err.Error())
	}
	err = cacheService.SetExpTime(pipelineId, expTime)
	if err != nil {
		grpclog.Error("RunCode: cache.SetExpTime: " + err.Error())
		return nil, errors.InternalError("Run code", "Error during set expiration to cache: "+err.Error())
	}

	go runCode(lc, exec, pipelineId)

	pipelineInfo := pb.RunCodeResponse{PipelineUuid: pipelineId.String()}
	return &pipelineInfo, nil
}

//CheckStatus is checking status for the specific pipeline by PipelineUuid
func (controller *playgroundController) CheckStatus(ctx context.Context, info *pb.CheckStatusRequest) (*pb.CheckStatusResponse, error) {
	// TODO implement this method
	status := pb.CheckStatusResponse{Status: pb.Status_STATUS_FINISHED}
	return &status, nil
}

//GetRunOutput is returning output of execution for specific pipeline by PipelineUuid
func (controller *playgroundController) GetRunOutput(ctx context.Context, info *pb.GetRunOutputRequest) (*pb.GetRunOutputResponse, error) {
	// TODO implement this method
	pipelineResult := pb.GetRunOutputResponse{Output: "Test Pipeline Result"}

	return &pipelineResult, nil
}

//GetCompileOutput is returning output of compilation for specific pipeline by PipelineUuid
func (controller *playgroundController) GetCompileOutput(ctx context.Context, info *pb.GetCompileOutputRequest) (*pb.GetCompileOutputResponse, error) {
	pipelineId := info.PipelineUuid
	compileOutputInterface, err := cacheService.GetValue(uuid.MustParse(pipelineId), cache.SubKey_CompileOutput)
	if err != nil {
		grpclog.Error("GetCompileOutput: cache.GetValue: error: " + err.Error())
		return nil, errors.NotFoundError("GetCompileOutput", "there is no compile output for pipelineId: "+pipelineId+", subKey: cache.SubKey_CompileOutput")
	}
	compileOutput, converted := compileOutputInterface.(string)
	if !converted {
		return nil, errors.InternalError("GetCompileOutput", "compile output can't be converted to string")
	}
	compileOutputResponse := pb.GetCompileOutputResponse{Output: compileOutput}

	return &compileOutputResponse, nil
}

// runCode validates, compiles and runs code by pipelineId.
// During each operation updates status of execution and saves it into cache.
// In case of compilation failed saves logs to cache.
// After success code running saves output to cache.
func runCode(lc *fs_tool.LifeCycle, exec *executors.Executor, pipelineId uuid.UUID) {
	// validate
	grpclog.Info("runCode: Validate ...")
	err := exec.Validate()
	if err != nil {
		// error during validation
		grpclog.Error("runCode: Validate: " + err.Error())

		// set to cache pipelineId: cache.SubKey_Status: pb.Status_STATUS_ERROR
		err = cacheService.SetValue(pipelineId, cache.SubKey_Status, pb.Status_STATUS_ERROR)
		if err != nil {
			grpclog.Error("runCode: cache.SetValue: " + err.Error())
			return
		}
	}
	grpclog.Info("runCode: Validate finish")

	grpclog.Info("runCode: Compile ...")
	err = exec.Compile()
	if err != nil {
		// error during compilation
		grpclog.Error("runCode: Compile: " + err.Error())

		// set to cache pipelineId: cache.SubKey_Status: pb.Status_STATUS_ERROR
		err = cacheService.SetValue(pipelineId, cache.SubKey_Status, pb.Status_STATUS_ERROR)
		if err != nil {
			grpclog.Error("runCode: cache.SetValue: " + err.Error())
			return
		}

		// set to cache pipelineId: cache.SubKey_CompileOutput: err.Error()
		err = cacheService.SetValue(pipelineId, cache.SubKey_CompileOutput, err.Error())
		if err != nil {
			grpclog.Error("runCode: cache.SetValue: " + err.Error())
			return
		}
	} else {
		// compilation success
		grpclog.Info("runCode: Compile finish")

		// set to cache pipelineId: cache.SubKey_CompileOutput: ""
		err = cacheService.SetValue(pipelineId, cache.SubKey_CompileOutput, "")
		if err != nil {
			grpclog.Error("runCode: cache.SetValue: " + err.Error())
			return
		}

		grpclog.Info("runCode: get executable file name ...")
		fileName, err := lc.GetExecutableName()
		if err != nil {
			grpclog.Error("runCode: get executable file name: " + err.Error())
			err = cacheService.SetValue(pipelineId, cache.SubKey_Status, pb.Status_STATUS_ERROR)
			if err != nil {
				grpclog.Error("runCode: cache.SetValue: " + err.Error())
				return
			}
		}
		grpclog.Infof("runCode: executable file name: %s", fileName)

		grpclog.Info("runCode: Run ...")
		output, err := exec.Run(fileName)
		if err != nil {
			// error during run code
			grpclog.Error("RunCode: Run: " + err.Error())

			// set to cache pipelineId: cache.SubKey_RunOutput: err.Error()
			err = cacheService.SetValue(pipelineId, cache.Subkey_RunOutput, err.Error())
			if err != nil {
				grpclog.Error("runCode: cache.SetValue: " + err.Error())
				return
			}

			// set to cache pipelineId: cache.SubKey_Status: pb.Status_STATUS_ERROR
			err = cacheService.SetValue(pipelineId, cache.SubKey_Status, pb.Status_STATUS_ERROR)
			if err != nil {
				grpclog.Error("runCode: cache.SetValue: " + err.Error())
				return
			}
		} else {
			// run code success
			grpclog.Info("runCode: Run finish")

			// set to cache pipelineId: cache.SubKey_RunOutput: output
			err = cacheService.SetValue(pipelineId, cache.Subkey_RunOutput, output)
			if err != nil {
				grpclog.Error("runCode: cache.SetValue: " + err.Error())
				return
			}

			// set to cache pipelineId: cache.SubKey_Status: pb.Status_STATUS_FINISHED
			err = cacheService.SetValue(pipelineId, cache.SubKey_Status, pb.Status_STATUS_FINISHED)
			if err != nil {
				grpclog.Error("runCode: cache.SetValue: " + err.Error())
				return
			}
		}
	}
	grpclog.Info("runCode: DeleteCompiledFile ...")
	err = lc.DeleteCompiledFile()
	if err != nil {
		grpclog.Error("RunCode: DeleteCompiledFile: " + err.Error())
	}
	grpclog.Info("runCode: DeleteExecutableFile ...")
	err = lc.DeleteExecutableFile()
	if err != nil {
		grpclog.Error("RunCode: DeleteExecutableFile: " + err.Error())
	}
	grpclog.Info("runCode: DeleteFolders ...")
	err = lc.DeleteFolders()
	if err != nil {
		grpclog.Error("RunCode: DeleteFolders: " + err.Error())
	}
}
