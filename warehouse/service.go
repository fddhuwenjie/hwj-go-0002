package warehouse

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
)

// Service owns inventory and committed order reservations.
type Service struct {
	mu           sync.Mutex
	stock        map[string]int
	reservations map[string]Reservation
}

// NewService creates an inventory service from an initial stock snapshot.
func NewService(initial map[string]int) (*Service, error) {
	stock := make(map[string]int, len(initial))
	for rawSKU, quantity := range initial {
		sku := normalizeSKU(rawSKU)
		if sku == "" || quantity < 0 {
			return nil, fmt.Errorf("%w: SKU %q has quantity %d", ErrInvalidInventory, rawSKU, quantity)
		}
		if _, duplicate := stock[sku]; duplicate {
			return nil, fmt.Errorf("%w: duplicate normalized SKU %s", ErrInvalidInventory, sku)
		}
		stock[sku] = quantity
	}
	return &Service{stock: stock, reservations: make(map[string]Reservation)}, nil
}

// ReserveOrder atomically validates and commits all allocations in an order.
func (s *Service) ReserveOrder(orderID string, lines []Line) (Reservation, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return Reservation{}, fmt.Errorf("%w: order ID is empty", ErrInvalidOrder)
	}
	normalized, err := normalizeLines(lines)
	if err != nil {
		return Reservation{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.reservations[orderID]; exists {
		return Reservation{}, fmt.Errorf("%w: %s", ErrOrderExists, orderID)
	}
	for _, line := range normalized {
		available, known := s.stock[line.SKU]
		if !known {
			return Reservation{}, fmt.Errorf("%w: %s", ErrUnknownSKU, line.SKU)
		}
		if available < line.Quantity {
			return Reservation{}, fmt.Errorf("%w: %s needs %d but has %d", ErrInsufficientStock, line.SKU, line.Quantity, available)
		}
	}
	for _, line := range normalized {
		s.stock[line.SKU] -= line.Quantity
	}

	reservation := Reservation{OrderID: orderID, Lines: cloneLines(normalized)}
	s.reservations[orderID] = reservation
	return cloneReservation(reservation), nil
}

// Reservation returns a previously committed order reservation.
func (s *Service) Reservation(orderID string) (Reservation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	reservation, ok := s.reservations[strings.TrimSpace(orderID)]
	return cloneReservation(reservation), ok
}

// Snapshot returns stock levels sorted by SKU.
func (s *Service) Snapshot() []StockLevel {
	s.mu.Lock()
	defer s.mu.Unlock()
	levels := make([]StockLevel, 0, len(s.stock))
	for sku, available := range s.stock {
		levels = append(levels, StockLevel{SKU: sku, Available: available})
	}
	sort.Slice(levels, func(i, j int) bool { return levels[i].SKU < levels[j].SKU })
	return levels
}

func normalizeLines(lines []Line) ([]Line, error) {
	if len(lines) == 0 {
		return nil, fmt.Errorf("%w: order has no lines", ErrInvalidOrder)
	}
	positions := make(map[string]int, len(lines))
	normalized := make([]Line, 0, len(lines))
	for _, line := range lines {
		sku := normalizeSKU(line.SKU)
		if sku == "" || line.Quantity <= 0 {
			return nil, fmt.Errorf("%w: SKU %q has quantity %d", ErrInvalidOrder, line.SKU, line.Quantity)
		}
		if index, exists := positions[sku]; exists {
			if line.Quantity > math.MaxInt-normalized[index].Quantity {
				return nil, fmt.Errorf("%w: quantity overflow for %s", ErrInvalidOrder, sku)
			}
			normalized[index].Quantity += line.Quantity
			continue
		}
		positions[sku] = len(normalized)
		normalized = append(normalized, Line{SKU: sku, Quantity: line.Quantity})
	}
	return normalized, nil
}

func normalizeSKU(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func cloneLines(lines []Line) []Line {
	return append([]Line(nil), lines...)
}

func cloneReservation(reservation Reservation) Reservation {
	reservation.Lines = cloneLines(reservation.Lines)
	return reservation
}
