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

package environment

import (
	"fmt"
	"time"
)

// NetworkEnvs contains all environment variables that need to run server.
type NetworkEnvs struct {
	ip   string
	port int
}

// NewNetworkEnvs constructor for NetworkEnvs
func NewNetworkEnvs(ip string, port int) *NetworkEnvs {
	return &NetworkEnvs{ip: ip, port: port}
}

// Address returns concatenated ip and port through ':'
func (serverEnvs NetworkEnvs) Address() string {
	return fmt.Sprintf("%s:%d", serverEnvs.ip, serverEnvs.port)
}

//CacheEnvs contains all environment variables that needed to use cache
type CacheEnvs struct {
	// cacheType is type of cache (local/redis)
	cacheType string

	// address os where the cache is running
	address string

	// keyExpirationTime is expiration time for cache keys
	keyExpirationTime time.Duration
}

// CacheType returns cache type
func (ce *CacheEnvs) CacheType() string {
	return ce.cacheType
}

// Address returns address to connect to remote cache service
func (ce *CacheEnvs) Address() string {
	return ce.address
}

// KeyExpirationTime returns expiration time for cache keys
func (ce *CacheEnvs) KeyExpirationTime() time.Duration {
	return ce.keyExpirationTime
}

// NewCacheEnvs constructor for CacheEnvs
func NewCacheEnvs(cacheType, cacheAddress string, cacheExpirationTime time.Duration) *CacheEnvs {
	return &CacheEnvs{
		cacheType:         cacheType,
		address:           cacheAddress,
		keyExpirationTime: cacheExpirationTime,
	}
}

//ApplicationEnvs contains all environment variables that needed to run backend processes
type ApplicationEnvs struct {
	// workingDir is a root working directory of application
	workingDir string

	// cacheEnvs contains environments for cache
	cacheEnvs *CacheEnvs

	// pipelineExecuteTimeout is timeout for code processing
	pipelineExecuteTimeout time.Duration
}

// NewApplicationEnvs constructor for ApplicationEnvs
func NewApplicationEnvs(workingDir string, cacheEnvs *CacheEnvs, pipelineExecuteTimeout time.Duration) *ApplicationEnvs {
	return &ApplicationEnvs{
		workingDir:             workingDir,
		cacheEnvs:              cacheEnvs,
		pipelineExecuteTimeout: pipelineExecuteTimeout,
	}
}

// WorkingDir returns root working directory of application
func (ae *ApplicationEnvs) WorkingDir() string {
	return ae.workingDir
}

// CacheEnvs returns cache environments
func (ae *ApplicationEnvs) CacheEnvs() CacheEnvs {
	return *ae.cacheEnvs
}

// PipelineExecuteTimeout returns timeout for code processing
func (ae *ApplicationEnvs) PipelineExecuteTimeout() time.Duration {
	return ae.pipelineExecuteTimeout
}
