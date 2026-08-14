package orderjson

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fddhuwenjie/hwj-go-0002/warehouse"
)

// Order is the JSON-facing representation of one reservation request.
type Order struct {
	OrderID string           `json:"order_id"`
	Lines   []warehouse.Line `json:"lines"`
}

type inventoryDocument struct {
	Stock map[string]int `json:"stock"`
}

// DecodeInventory reads one inventory JSON document.
func DecodeInventory(source io.Reader) (map[string]int, error) {
	var document inventoryDocument
	decoder := json.NewDecoder(source)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return nil, fmt.Errorf("decode inventory: %w", err)
	}
	if document.Stock == nil {
		return nil, fmt.Errorf("decode inventory: stock is required")
	}
	return document.Stock, nil
}

// DecodeOrder reads one order JSON document.
func DecodeOrder(source io.Reader) (Order, error) {
	var order Order
	decoder := json.NewDecoder(source)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&order); err != nil {
		return Order{}, fmt.Errorf("decode order: %w", err)
	}
	return order, nil
}

// Reserve decodes an order and applies it to the supplied service.
func Reserve(service *warehouse.Service, source io.Reader) (warehouse.Reservation, error) {
	order, err := DecodeOrder(source)
	if err != nil {
		return warehouse.Reservation{}, err
	}
	return service.ReserveOrder(order.OrderID, order.Lines)
}
