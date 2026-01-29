package services

import (
	pb "github.com/chazu/yutani/pkg/proto/yutani"
	"github.com/chazu/yutani/pkg/server"
	"google.golang.org/grpc"
)

// RegisterAllServices registers all 11 Yutani gRPC services on the given
// gRPC server, using srv as the backing Yutani server instance.
func RegisterAllServices(grpcServer *grpc.Server, srv *server.Server) {
	pb.RegisterSessionServiceServer(grpcServer, NewSessionService(srv))
	pb.RegisterScreenServiceServer(grpcServer, NewScreenService(srv))
	pb.RegisterEventServiceServer(grpcServer, NewEventService(srv))
	pb.RegisterWidgetServiceServer(grpcServer, NewWidgetService(srv))
	pb.RegisterListServiceServer(grpcServer, NewListService(srv))
	pb.RegisterTableServiceServer(grpcServer, NewTableService(srv))
	pb.RegisterFormServiceServer(grpcServer, NewFormService(srv))
	pb.RegisterTreeServiceServer(grpcServer, NewTreeService(srv))
	pb.RegisterLayoutServiceServer(grpcServer, NewLayoutService(srv))
	pb.RegisterDebugServiceServer(grpcServer, NewDebugService(srv))
	pb.RegisterTestServiceServer(grpcServer, NewTestService(srv))
}
