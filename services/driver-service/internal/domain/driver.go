package domain

import (
	pb "github.com/AuraReaper/voom/shared/proto/driver"
)

type Driver struct {
	ID             string
	Name           string
	PackageSlug    string
	CarPlate       string
	ProfilePicture string
	Location       *pb.Location
	Geohash        string
}

type DriverRepository interface {
	GetDrivers() ([]*Driver, error)
	RegisterDriver(driver *Driver) (*Driver, error)
	FindAvailableDrivers(packageType string) ([]string, error)
}
