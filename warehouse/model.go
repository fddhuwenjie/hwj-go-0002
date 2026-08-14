package warehouse

import "errors"

var (
	ErrInvalidInventory  = errors.New("invalid inventory")
	ErrInvalidOrder      = errors.New("invalid order")
	ErrOrderExists       = errors.New("order already reserved")
	ErrUnknownSKU        = errors.New("unknown SKU")
	ErrInsufficientStock = errors.New("insufficient stock")
)

// Line describes one requested stock allocation.
type Line struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

// Reservation is the normalized allocation committed for an order.
type Reservation struct {
	OrderID string `json:"order_id"`
	Lines   []Line `json:"lines"`
}

// StockLevel is one normalized inventory entry.
type StockLevel struct {
	SKU       string `json:"sku"`
	Available int    `json:"available"`
}
