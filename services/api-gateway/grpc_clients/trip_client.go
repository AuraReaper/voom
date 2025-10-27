package grpc_clients

import (
	"os"

	pb "github.com/AuraReaper/voom/shared/proto/trip"
	"github.com/AuraReaper/voom/shared/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type tripServiceclient struct {
	Client pb.TripServiceClient
	conn   *grpc.ClientConn
}

func NewTripServiceClient() (*tripServiceclient, error) {
	tripServiceURL := os.Getenv("TRIP_SERVICE_URL ")
	if tripServiceURL == "" {
		tripServiceURL = "trip-service:9093"
	}

	dialOptions := append(
		tracing.DialOptionsWithTracing(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	conn, err := grpc.NewClient(tripServiceURL, dialOptions...)
	if err != nil {
		return nil, err
	}

	client := pb.NewTripServiceClient(conn)

	return &tripServiceclient{
		Client: client,
		conn:   conn,
	}, nil
}

func (c *tripServiceclient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
