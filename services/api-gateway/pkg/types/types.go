package types

import (
	pb "github.com/AuraReaper/voom/shared/proto/trip"
	"github.com/AuraReaper/voom/shared/types"
)

type PreviewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (p *PreviewTripRequest) ToProto() *pb.PreviewTripRequest {
	return &pb.PreviewTripRequest{
		UserID: p.UserID,
		StartLocation: &pb.Coordinate{
			Latitude:  p.Pickup.Latitude,
			Longitude: p.Pickup.Longitude,
		},
		EndLocation: &pb.Coordinate{
			Latitude:  p.Destination.Latitude,
			Longitude: p.Destination.Longitude,
		},
	}
}

type CreateTripRequest struct {
	RiderFairID string `json:"rideFareID"`
	UserID      string `json:"userID"`
}

func (p *CreateTripRequest) ToProto() *pb.CreateTripRequest {
	return &pb.CreateTripRequest{
		RideFareID: p.RiderFairID,
		UserID:     p.UserID,
	}
}
