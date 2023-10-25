package handlers

import (
	"context"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"go.uber.org/zap"
)

type ScoringSystemHandler struct {
	scoringRepository database.ScoringRepository
}

func NewScoringSystem(scoringRepository database.ScoringRepository) *ScoringSystemHandler {
	return &ScoringSystemHandler{
		scoringRepository: scoringRepository,
	}
}

func (h *ScoringSystemHandler) GetOrderStatus(ctx context.Context) (string, error) {

	orderID, err := h.scoringRepository.GetOrderStatus(ctx)
	if err != nil {
		logger.Log.Error("Error order exist", zap.Error(err))
	}

	return orderID, nil
}

func (h *ScoringSystemHandler) UpdateOrder(ctx context.Context, order models.ScoringSystem) error {
	err := h.scoringRepository.UpdateOrder(ctx, order)
	if err != nil {

		return err
	}
	return nil
}
