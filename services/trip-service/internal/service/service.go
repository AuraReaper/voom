package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AuraReaper/voom/services/trip-service/internal/domain"
	tripTypes "github.com/AuraReaper/voom/services/trip-service/pkg/types"
	pbd "github.com/AuraReaper/voom/shared/proto/driver"
	"github.com/AuraReaper/voom/shared/proto/trip"
	"github.com/AuraReaper/voom/shared/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripService struct {
	repo domain.TripRepository
}

func NewTripService(repo domain.TripRepository) *TripService {
	return &TripService{
		repo: repo,
	}
}

func (s *TripService) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
		Driver:   &trip.TripDriver{},
	}

	return s.repo.CreateTrip(ctx, t)
}

func (s *TripService) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude, pickup.Latitude,
		destination.Longitude, destination.Latitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route OSRM API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the respone: %v", err)
	}

	var routeResp tripTypes.OsrmApiResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &routeResp, nil
}

func (s *TripService) EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()
	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, f := range baseFares {
		estimatedFares[i] = estimateRouteFare(f, route)
	}

	return estimatedFares
}

func (s *TripService) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFareModel, userID string, route *tripTypes.OsrmApiResponse) ([]*domain.RideFareModel, error) {
	fares := make([]*domain.RideFareModel, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()

		fare := &domain.RideFareModel{
			UserID:          userID,
			ID:              id,
			PackageSlug:     f.PackageSlug,
			TotalPriceInINR: f.TotalPriceInINR,
			Route:           route,
		}

		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save trip fare: %s", err)
		}

		fares[i] = fare
	}

	return fares, nil
}

func (s *TripService) GetAndValidateFare(ctx context.Context, fareID, userID string) (*domain.RideFareModel, error) {
	fare, err := s.repo.GetRideFareByID(ctx, fareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ride fair: %w", err)
	}

	if fare == nil {
		return nil, fmt.Errorf("fare does not exsist")
	}

	if userID != fare.UserID {
		return nil, fmt.Errorf("fare does not  belong to the user")
	}

	return fare, nil
}

func estimateRouteFare(f *domain.RideFareModel, route *tripTypes.OsrmApiResponse) *domain.RideFareModel {
	pricingCfg := tripTypes.DefaultPricingConfig()
	PackagePrice := f.TotalPriceInINR

	distanceKm := route.Route[0].Distance / 1000
	durationInMin := route.Route[0].Duration / 60

	distanceFare := distanceKm * pricingCfg.PricePerKm
	timeFare := durationInMin * pricingCfg.PricePerMinute

	totalFare := PackagePrice + distanceFare + timeFare

	return &domain.RideFareModel{
		TotalPriceInINR: totalFare,
		PackageSlug:     f.PackageSlug,
	}
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:     "suv",
			TotalPriceInINR: 150,
		},
		{
			PackageSlug:     "sedan",
			TotalPriceInINR: 100,
		},
		{
			PackageSlug:     "van",
			TotalPriceInINR: 200,
		},
		{
			PackageSlug:     "luxury",
			TotalPriceInINR: 500,
		},
	}
}

func (s *TripService) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error) {
	return s.repo.GetTripByID(ctx, id)
}

func (s *TripService) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error {
	return s.repo.UpdateTrip(ctx, tripID, status, driver)
}
