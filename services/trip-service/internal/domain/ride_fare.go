package domain

import (
	tripTypes "github.com/AuraReaper/voom/services/trip-service/pkg/types"
	pb "github.com/AuraReaper/voom/shared/proto/trip"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID              primitive.ObjectID
	UserID          string
	PackageSlug     string // ex: van, luxury, sedan
	TotalPriceInINR float64
	Route           *tripTypes.OsrmApiResponse
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:              r.ID.Hex(),
		UserID:          r.UserID,
		PackageSlug:     r.PackageSlug,
		TotalPriceInINR: r.TotalPriceInINR,
	}
}

func ToRidesFaresProto(fares []*RideFareModel) []*pb.RideFare {
	var protoFares []*pb.RideFare
	for _, f := range fares {
		protoFares = append(protoFares, f.ToProto())
	}

	return protoFares
}
