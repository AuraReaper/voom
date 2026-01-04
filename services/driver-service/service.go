package main

import (
	"log"
	"math/rand"

	"github.com/AuraReaper/voom/services/driver-service/internal/domain"
	pb "github.com/AuraReaper/voom/shared/proto/driver"
	"github.com/AuraReaper/voom/shared/util"
	"github.com/mmcloughlin/geohash"
)

type Service struct {
	repo domain.DriverRepository
}

func NewService(repo domain.DriverRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetDrivers() ([]*pb.Driver, error) {
	drivers, err := s.repo.GetDrivers()
	if err != nil {
		return nil, err
	}

	log.Printf("Returning %d drivers", len(drivers))

	var pbDrivers []*pb.Driver
	for _, d := range drivers {
		pbDrivers = append(pbDrivers, &pb.Driver{
			Id:             d.ID,
			Name:           d.Name,
			PackageSlug:    d.PackageSlug,
			CarPlate:       d.CarPlate,
			ProfilePicture: d.ProfilePicture,
			Location:       d.Location,
			Geohash:        d.Geohash,
		})
	}

	return pbDrivers, nil
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	newDriver := &domain.Driver{
		ID:             driverId,
		Name:           "New Driver",
		PackageSlug:    packageSlug,
		CarPlate:       "NEW-1234",
		ProfilePicture: util.GetRandomAvatar(int(rand.Int31n(100))),
		Location:       &pb.Location{Latitude: 20.27, Longitude: 85.82},
		Geohash:        geohash.Encode(20.27, 85.82),
	}

	d, err := s.repo.RegisterDriver(newDriver)
	if err != nil {
		return nil, err
	}

	return &pb.Driver{
		Id:             d.ID,
		Name:           d.Name,
		PackageSlug:    d.PackageSlug,
		CarPlate:       d.CarPlate,
		ProfilePicture: d.ProfilePicture,
		Location:       d.Location,
		Geohash:        d.Geohash,
	}, nil
}

func (s *Service) UnregisterDriver(driverId string) {
	// No-op
}

func (s *Service) FindAvailableDrivers(packageType string) ([]string, error) {
	return s.repo.FindAvailableDrivers(packageType)
}
