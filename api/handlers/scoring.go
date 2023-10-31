package handlers

import (
	"context"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"go.uber.org/zap"
)

type ScoringSystem struct {
	scoringRepo database.ScoringRepository
}

func NewScoringSystem(repo database.ScoringRepository) *ScoringSystem {
	return &ScoringSystem{
		scoringRepo: repo,
	}
}

func (h *ScoringSystem) GetOrderStatus(ctx context.Context) (string, error) {

	orderID, err := h.scoringRepo.GetOrderStatus(ctx)
	if err != nil {

		logger.Log.Error("Error order exist", zap.Error(err))
	}

	return orderID, nil
}

func (h *ScoringSystem) UpdateOrder(ctx context.Context, order models.ScoringSystem) error {
	err := h.scoringRepo.UpdateOrder(ctx, order)
	if err != nil {

		return err
	}
	return nil
}
