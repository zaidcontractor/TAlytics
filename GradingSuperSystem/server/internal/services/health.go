package services

import (
	"context"

	pb "github.com/talytics/server/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HealthService struct {
	pb.UnimplementedHealthServiceServer
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:    "TAlytics Go server is running",
		Timestamp: timestamppb.Now(),
	}, nil
}