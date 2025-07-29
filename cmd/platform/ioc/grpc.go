// Copyright 2025 ecodeclub
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ioc

import (
	chatv1 "github.com/ecodeclub/ai-gateway-go/api/proto/gen/chat/v1"
	igrpc "github.com/ecodeclub/ai-gateway-go/internal/grpc"
	"github.com/gotomicro/ego/server/egrpc"
)

func InitGrpcServer(chatSvc *igrpc.ChatServer) *egrpc.Component {
	grpcComponent := egrpc.Load("grpc.server").Build()
	chatv1.RegisterServiceServer(grpcComponent, chatSvc)
	return grpcComponent
}
