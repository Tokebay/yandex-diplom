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
	"net/http"
	"time"
)

type APIAccrualSystem struct {
	ScoringSystemHandler *handlers.ScoringSystemHandler
	Config               *config.Config
}

func (s *APIAccrualSystem) ScoringSystem() {
	for {
		// Получаем номер заказа из системы accrual
		orderID, err := s.ScoringSystemHandler.GetOrderStatus(context.Background())
		if err != nil {
			logger.Log.Error("Error getting order status", zap.Error(err))
			continue
		}

		// Создаем ссылку для запроса GET
		accAddr := fmt.Sprintf("%s/api/orders/%s", s.Config.AccrualSystemPort, orderID)
		resp, err := http.Get(accAddr)
		if err != nil {
			logger.Log.Error("Error GET request failed to accrual system", zap.Error(err))
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var orderScoring models.ScoringSystem
			if err := json.NewDecoder(resp.Body).Decode(&orderScoring); err != nil {
				logger.Log.Error("Error decoding response from scoring system", zap.Error(err))
			} else {
				// Обновляем данные заказа и начисляем бонусы
				if err := s.ScoringSystemHandler.UpdateOrder(context.Background(), orderScoring); err != nil {
					logger.Log.Error("Error updating order", zap.Error(err))
				}
			}
			// Закрываем тело ответа после успешной обработки
			resp.Body.Close()
		} else {
			// Закрываем тело ответа в случае ошибочного статуса
			resp.Body.Close()
		}

		// Ждем какой-то интервал перед следующим запросом к системе расчета баллов
		time.Sleep(time.Minute)
	}
}
