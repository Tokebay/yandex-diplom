package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Tokebay/yandex-diplom/api/logger"
	"github.com/Tokebay/yandex-diplom/database"
	"github.com/Tokebay/yandex-diplom/domain/models"
	"go.uber.org/zap"
)

var ErrUnauthorized = errors.New("user not auth")

//var ErrNoContent = errors.New("no content")

type OrderHandler struct {
	orderRepository database.OrderRepository
}

func NewOrderHandler(orderRepository database.OrderRepository) *OrderHandler {
	return &OrderHandler{
		orderRepository: orderRepository,
	}
}

func (h *OrderHandler) UploadOrderHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	fmt.Println("UploadOrderHandler")

	// Получил userID из куки
	userID, err := GetUserCookie(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// получил номера заказа из тела запроса
	orderNumber, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Printf("orderNumber %s; userID %d \n", orderNumber, userID)

	isValidOrder := isValidLuhnAlgorithm(string(orderNumber))
	if !isValidOrder {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// был ли номер заказа уже загружен этим пользователем
	isOrderExist, err := h.orderRepository.OrderExists(userID, string(orderNumber))
	if err != nil {
		logger.Log.Error("Error order exist", zap.Error(err))
	}

	if isOrderExist {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Проверю что номер заказа не был загружен другим пользователем
	isOrderUploaded, err := h.orderRepository.OrderExistsByNumber(string(orderNumber))
	if err != nil {
		logger.Log.Error("Error order uploaded by another", zap.Error(err))
	}
	if isOrderUploaded {
		w.WriteHeader(http.StatusConflict)
		return
	}

	order := models.Order{
		Number:    string(orderNumber),
		Status:    models.OrderStatusNew,
		CreatedAt: time.Now().Format(time.RFC3339),
		Accrual:   0,
		UserID:    userID,
	}

	// Сохраняю номера заказа в БД со статусом  - NEW
	err = h.orderRepository.UploadOrder(order)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Новый номер заказа принят в обработку
	w.WriteHeader(http.StatusAccepted)
}

// GetOrdersHandler Номера заказа в выдаче отсортированы по времени загрузки от самых старых к самым новым. Формат даты — RFC3339.
func (h *OrderHandler) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetOrdersHandler")

	//orders, err := getOrders(r.Context(), h)
	//if err != nil {
	//	if errors.Is(err, database.ErrDataNotFound) {
	//		logger.Log.Error("Error No Content", zap.Error(err))
	//		w.WriteHeader(http.StatusNoContent)
	//		return
	//	}
	//
	//	if errors.Is(err, ErrUnauthorized) {
	//		logger.Log.Error("Error Unauthorized", zap.Error(err))
	//		w.WriteHeader(http.StatusUnauthorized)
	//		return
	//	}
	//
	//	logger.Log.Error("Error getting orders", zap.Error(err))
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	userID, err := GetUserCookie(r)
	if err != nil && userID == -1 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	//Получение списка заказов для пользователя из базы данных
	orders, err := h.orderRepository.GetOrdersByUserID(userID)
	if err != nil {
		if errors.Is(err, database.ErrDataNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		logger.Log.Error("Error getting orders", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Преобразование списка заказов в JSON
	ordersResp, err := json.Marshal(orders)
	if err != nil {
		logger.Log.Error("Error marshaling JSON", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Отправка успешного ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(ordersResp)
}

func getOrders(ctx context.Context, h *OrderHandler) ([]models.OrderResponse, error) {
	userID, ok := ctx.Value(models.UserIDKey).(int)
	fmt.Printf("getOrders. userID %d \n", userID)
	if !ok {
		return nil, errors.New("incorrect user id")
	}

	orders, err := h.orderRepository.GetOrdersByUserID(userID)
	if err != nil {
		logger.Log.Error("Error getting orders", zap.Error(err))
		return nil, err
	}

	return orders, nil
}

func isValidLuhnAlgorithm(number string) bool {

	// Проверка, что номер заказа состоит только из цифр
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Проверка, что номер заказа имеет правильную длину (не менее 2 цифр)
	if len(number) < 2 {
		return false
	}

	digits := make([]int, len(number))
	for i, char := range number {
		digit, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}
		digits[len(digits)-1-i] = digit
	}

	sum := 0
	double := false
	for _, digit := range digits {
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}

	return sum%10 == 0
}
