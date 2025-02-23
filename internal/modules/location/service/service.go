package service

import (
	"context"
	"math"
	"strconv"

	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/modules/location"
	"github.com/yaroslavvasilenko/argon/internal/modules/location/storage"
)

type Location struct {
	s      *storage.Location
	logger *logger.Glog
}

func NewLocation(s *storage.Location, logger *logger.Glog) *Location {
	srv := &Location{
		s:      s,
		logger: logger,
	}

	return srv
}


func (l *Location) GetLocation(ctx context.Context, req location.GetLocationRequest) (location.GetLocationResponse, error) {
	locResp, err := l.s.GetLocation(ctx, req.Area.Coordinates.Lat, req.Area.Coordinates.Lng, calculateZoomForRadius(req.Area.Radius, req.Area.Coordinates.Lat), req.Language)
	if err != nil {
		return location.GetLocationResponse{}, err
	}

	response := location.GetLocationResponse{
		ID:   strconv.FormatInt(locResp.PlaceID, 10),
		Name: locResp.DisplayName,
		Area: req.Area,
	}

	return response, nil
}

func calculateZoomForRadius(radius int, latitude float64) int {
	// Mercator projection scale factor at given latitude
	latRad := ((math.Abs(latitude) + 1) * math.Pi) / 180
	scale := 1 / math.Cos(latRad)

	// Adjust radius based on latitude to compensate for projection distortion
	adjustedRadius := float64(radius) * scale

	// Base zoom levels calibrated for equator
	switch {
	case adjustedRadius <= 150:
		return 18
	case adjustedRadius <= 300:
		return 17
	case adjustedRadius <= 600:
		return 16
	case adjustedRadius <= 1200:
		return 15
	case adjustedRadius <= 2500:
		return 14
	case adjustedRadius <= 5000:
		return 13
	case adjustedRadius <= 10000:
		return 12
	case adjustedRadius <= 20000:
		return 11
	case adjustedRadius <= 40000:
		return 10
	case adjustedRadius <= 80000:
		return 9
	case adjustedRadius <= 160000:
		return 8
	case adjustedRadius <= 320000:
		return 7
	case adjustedRadius <= 640000:
		return 6
	case adjustedRadius <= 1280000:
		return 5
	case adjustedRadius <= 2560000:
		return 4
	case adjustedRadius <= 5120000:
		return 3
	case adjustedRadius <= 10240000:
		return 2
	case adjustedRadius <= 20480000:
		return 1
	default:
		return 0
	}
}
