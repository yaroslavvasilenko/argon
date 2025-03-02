package service

import (
	"context"

	"github.com/yaroslavvasilenko/argon/internal/core/logger"
	"github.com/yaroslavvasilenko/argon/internal/modules/boost/storage"
	"github.com/yaroslavvasilenko/argon/internal/models"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/internal/modules/boost"
)

type Boost struct {
	s      *storage.Boost
	logger *logger.Glog
}

func NewBoost(s *storage.Boost, logger *logger.Glog) *Boost {
	srv := &Boost{
		s:      s,
		logger: logger,
	}

	return srv
}

func (s *Boost) GetBoost(ctx context.Context, id uuid.UUID) (boost.GetBoostResponse, error) {
	boosts, err := s.s.GetBoosts(ctx, id)
	if err != nil {
		return boost.GetBoostResponse{}, err
	}

	resp := boost.GetBoostResponse{}

	for _, b := range boosts {
		resp.EnableBoostTypes = append(resp.EnableBoostTypes, b.Type)
	}

	for boostType, commission := range models.GetBoostTypesWithCommissions() {
		resp.BoostResp = append(resp.BoostResp, boost.BoostResp{
			Type:              boostType,
			CommissionPercent: commission,
		})
	}



	return resp, nil
}

func (s *Boost) UpsertBoost(ctx context.Context, req boost.UpdateBoostRequest) (boost.GetBoostResponse, error) {
	boosts := make([]models.Boost, 0, len(req.Boosts))
	
	for _, t := range req.Boosts {
		boosts = append(boosts, models.Boost{
			ListingID: req.ListingID,
			Type:  t,
		})
	}
	
	
	err := s.s.UpsertBoost(ctx, req.ListingID, boosts)
	if err != nil {
		return boost.GetBoostResponse{}, err
	}
	
	return s.GetBoost(ctx, req.ListingID)
}