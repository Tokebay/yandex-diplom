package accrual

import (
	"context"
	"encoding/json"
	"github.com/Tokebay/yandex-diplom/api/handlers"
	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/config"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type APIAccrualSystem struct {
	ScoringSystemHandler *handlers.ScoringSystemHandler
	Config               *config.Config
}

func (a *APIAccrualSystem) ScoringSystem(done chan struct{}, ctx context.Context) {
	for {
		// Получаем заказ со статусом NEW
		orderID, err := a.ScoringSystemHandler.GetOrderStatus(context.Background())
		if err != nil {
			logger.Log.Error("Error getting order status", zap.Error(err))
			return
		}

		// делаем http GET запрос в AccrualSystem
		orderScoring, err := GetHTTPRequest(ctx, orderID, a.Config)
		if err != nil {
			logger.Log.Error("Error GET request failed to accrual system", zap.Error(err))
			return
		}
		if err := a.ScoringSystemHandler.UpdateOrder(context.Background(), *orderScoring); err != nil {
			logger.Log.Error("Error updating order", zap.Error(err))
		}

		// Ждем какой-то интервал перед следующим запросом к системе расчета баллов
		select {
		case <-time.After(100 * time.Millisecond):
		case <-done:
			return
		}
	}
}

func GetHTTPRequest(ctx context.Context, orderNum string, cfg *config.Config) (*models.ScoringSystem, error) {
	var order models.ScoringSystem
	client := &http.Client{
		Timeout: time.Second * 20,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", cfg.AccrualSystemAddr+"/api/orders/"+orderNum, nil)
	//req, err := http.NewRequestWithContext(ctx, "GET", accAddress, nil)
	if err != nil {
		logger.Log.Error("error create req", zap.Error(err))
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		logger.Log.Error("error send http req", zap.Error(err))
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		logger.Log.Error("HTTP request status", zap.Error(err))
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	defer response.Body.Close()

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &order)
	if err != nil {
		logger.Log.Error("error unmarshal body", zap.Error(err))
		return nil, err
	}

	return &order, nil
}
