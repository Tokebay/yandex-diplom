package accrual

import (
	"context"
	"encoding/json"
	"fmt"
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
	ScoringSystemHandler *handlers.ScoringSystem
	Config               *config.Config
}

func (a *APIAccrualSystem) ScoringSystem() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Используйте контекст с длительным таймаутом
	defer cancel()
	// Получаем заказ со статусом NEW
	orderID, err := a.ScoringSystemHandler.GetOrderStatus(ctx)
	if err != nil {
		logger.Log.Error("Error getting order status", zap.Error(err))
		return
	}

	// делаем http GET запрос в AccrualSystem
	orderScoring, err := GetHTTPRequest(ctx, orderID, a.Config)
	if err != nil {
		logger.Log.Error("HTTP GET request failed to accrual system", zap.Error(err))
		return
	}

	if err := a.ScoringSystemHandler.UpdateOrder(ctx, *orderScoring); err != nil {
		logger.Log.Error("Error updating order", zap.Error(err))
	}

}

func GetHTTPRequest(ctx context.Context, orderNum string, cfg *config.Config) (*models.ScoringSystem, error) {
	var order models.ScoringSystem
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	URI := cfg.AccrualSystemAddr + "/api/orders/" + orderNum
	//fmt.Printf("accrual URI %s \n", URI)

	req, err := http.NewRequestWithContext(ctx, "GET", URI, nil)
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
		err := fmt.Errorf("unexpected HTTP status: %s", response.Status)
		logger.Log.Error("HTTP request status", zap.Error(err))
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Log.Error("error ReadAll", zap.Error(err))
		return nil, err
	}
	defer response.Body.Close()

	err = json.Unmarshal(body, &order)
	if err != nil {
		logger.Log.Error("error unmarshal body", zap.Error(err))
		return nil, err
	}
	// Логируем значение структуры order
	orderLogger := zap.Any("order", order)
	logger.Log.Info("Received order details", orderLogger)

	return &order, nil
}
