package repository

import (
	"log"

	"github.com/AuraReaper/voom/services/driver-service/internal/domain"
	
)

type inmemDriverRepository struct {
	drivers []*domain.Driver
}

func NewInmemDriverRepository() domain.DriverRepository {
	return &inmemDriverRepository{
		drivers: []*domain.Driver{},
	}
}

func (r *inmemDriverRepository) GetDrivers() ([]*domain.Driver, error) {
	log.Printf("Returning %d drivers", len(r.drivers))
	return r.drivers, nil
}

func (r *inmemDriverRepository) RegisterDriver(driver *domain.Driver) (*domain.Driver, error) {
	for _, d := range r.drivers {
		if d.ID == driver.ID {
			return d, nil
		}
	}

	r.drivers = append(r.drivers, driver)

	return driver, nil
}

func (r *inmemDriverRepository) FindAvailableDrivers(packageType string) ([]string, error) {
	var matchingDrivers []string

	for _, driver := range r.drivers {
		if driver.PackageSlug == packageType {
			matchingDrivers = append(matchingDrivers, driver.ID)
		}
	}

	if len(matchingDrivers) == 0 {
		return []string{}, nil
	}

	return matchingDrivers, nil
}

/*
var defaultDrivers = []*pb.Driver{
	{
		Id:             "driver-1",
		Name:           "Suresh Kumar",
		PackageSlug:    "sedan",
		CarPlate:       "OD02AB1234",
		ProfilePicture: util.GetRandomAvatar(1),
		Location:       &pb.Location{Latitude: 20.344744, Longitude: 85.803818},
		Geohash:        geohash.Encode(20.344744, 85.803818),
	},
	{
		Id:             "driver-2",
		Name:           "Ramesh Sahoo",
		PackageSlug:    "suv",
		CarPlate:       "OD02CD5678",
		ProfilePicture: util.GetRandomAvatar(2),
		Location:       &pb.Location{Latitude: 20.3533, Longitude: 85.8253},
		Geohash:        geohash.Encode(20.3533, 85.8253),
	},
	{
		Id:             "driver-3",
		Name:           "Amit Sharma",
		PackageSlug:    "van",
		CarPlate:       "OD02EF9012",
		ProfilePicture: util.GetRandomAvatar(3),
		Location:       &pb.Location{Latitude: 20.2883, Longitude: 85.8435},
		Geohash:        geohash.Encode(20.2883, 85.8435),
	},
	{
		Id:             "driver-4",
		Name:           "Priya Das",
		PackageSlug:    "luxury",
		CarPlate:       "OD02GH3456",
		ProfilePicture: util.GetRandomAvatar(4),
		Location:       &pb.Location{Latitude: 20.24, Longitude: 85.83},
		Geohash:        geohash.Encode(20.24, 85.83),
	},
	{
		Id:             "driver-5",
		Name:           "Sunita Mohanty",
		PackageSlug:    "sedan",
		CarPlate:       "OD02IJ7890",
		ProfilePicture: util.GetRandomAvatar(5),
		Location:       &pb.Location{Latitude: 20.26, Longitude: 85.77},
		Geohash:        geohash.Encode(20.26, 85.77),
	},
	{
		Id:             "driver-6",
		Name:           "Manoj Swain",
		PackageSlug:    "suv",
		CarPlate:       "OD02KL1234",
		ProfilePicture: util.GetRandomAvatar(6),
		Location:       &pb.Location{Latitude: 20.30, Longitude: 85.82},
		Geohash:        geohash.Encode(20.30, 85.82),
	},
	{
		Id:             "driver-7",
		Name:           "Rakesh Kumar",
		PackageSlug:    "sedan",
		CarPlate:       "BR01AB1234",
		ProfilePicture: util.GetRandomAvatar(7),
		Location:       &pb.Location{Latitude: 25.253391, Longitude: 86.989059},
		Geohash:        geohash.Encode(25.253391, 86.989059),
	},
	{
		Id:             "driver-8",
		Name:           "Sanjay Sahoo",
		PackageSlug:    "suv",
		CarPlate:       "BR01CD5678",
		ProfilePicture: util.GetRandomAvatar(8),
		Location:       &pb.Location{Latitude: 25.25, Longitude: 86.98},
		Geohash:        geohash.Encode(25.25, 86.98),
	},
	{
		Id:             "driver-9",
		Name:           "Vijay Sharma",
		PackageSlug:    "van",
		CarPlate:       "BR01EF9012",
		ProfilePicture: util.GetRandomAvatar(9),
		Location:       &pb.Location{Latitude: 25.26, Longitude: 86.99},
		Geohash:        geohash.Encode(25.26, 86.99),
	},
	{
		Id:             "driver-10",
		Name:           "Anita Das",
		PackageSlug:    "luxury",
		CarPlate:       "BR01GH3456",
		ProfilePicture: util.GetRandomAvatar(10),
		Location:       &pb.Location{Latitude: 25.24, Longitude: 86.97},
		Geohash:        geohash.Encode(25.24, 86.97),
	},
	{
		Id:             "driver-11",
		Name:           "Sunil Mohanty",
		PackageSlug:    "sedan",
		CarPlate:       "BR01IJ7890",
		ProfilePicture: util.GetRandomAvatar(11),
		Location:       &pb.Location{Latitude: 25.27, Longitude: 86.96},
		Geohash:        geohash.Encode(25.27, 86.96),
	},
}
*/
