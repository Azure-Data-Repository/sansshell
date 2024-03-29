/* Copyright (c) 2019 Snowflake Inc. All rights reserved.

   Licensed under the Apache License, Version 2.0 (the
   "License"); you may not use this file except in compliance
   with the License.  You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing,
   software distributed under the License is distributed on an
   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
   KIND, either express or implied.  See the License for the
   specific language governing permissions and limitations
   under the License.
*/

package server

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Snowflake-Labs/sansshell/services/sansshell"
)

var (
	// Version is the value returned by the Version RPC service.
	// This should likely be set as a linker option from external
	// input such as a git SHA or RPM version number.
	// go build -ldflags="-X github.com/Snowflake-Labs/sansshell/services/sansshell/server.Version=..."
	//
	// NOTE: This is a var so the linker can set it but in reality it's a const so treat as such.
	Version string
)

func (s *Server) Version(ctx context.Context, req *emptypb.Empty) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{
		Version: Version,
	}, nil
}
