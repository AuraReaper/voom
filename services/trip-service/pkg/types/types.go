package types

import (
	pb "github.com/AuraReaper/voom/shared/proto/trip"
)

type OsrmApiResponse struct {
	Route []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes" `
}

func (o *OsrmApiResponse) ToProto() *pb.Route {
	if len(o.Route) == 0 {
		return &pb.Route{}
	}
	route := o.Route[0]
	geometry := route.Geometry.Coordinates
	coordinates := make([]*pb.Coordinate, len(geometry))

	for i, coord := range geometry {
		coordinates[i] = &pb.Coordinate{
			Latitude:  coord[0],
			Longitude: coord[1],
		}
	}

	return &pb.Route{
		Geometry: []*pb.Geometry{
			{
				Coordinates: coordinates,
			},
		},
		Distance: route.Distance,
		Duration: route.Duration,
	}
}

type PricingConfig struct {
	PricePerKm     float64
	PricePerMinute float64
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerKm:     12,
		PricePerMinute: 1,
	}
}
