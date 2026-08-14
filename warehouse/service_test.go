package warehouse_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/fddhuwenjie/hwj-go-0002/warehouse"
)

func TestReserveOrderAggregatesLinesAndUpdatesStock(t *testing.T) {
	service := mustService(t, map[string]int{"box-small": 10, "TAPE": 4})
	reservation, err := service.ReserveOrder("order-1", []warehouse.Line{
		{SKU: " box-small ", Quantity: 2},
		{SKU: "BOX-SMALL", Quantity: 3},
		{SKU: "tape", Quantity: 1},
	})
	if err != nil {
		t.Fatalf("ReserveOrder returned an error: %v", err)
	}
	wantLines := []warehouse.Line{{SKU: "BOX-SMALL", Quantity: 5}, {SKU: "TAPE", Quantity: 1}}
	if !reflect.DeepEqual(reservation.Lines, wantLines) {
		t.Fatalf("unexpected reservation lines: %#v", reservation.Lines)
	}
	wantStock := []warehouse.StockLevel{{SKU: "BOX-SMALL", Available: 5}, {SKU: "TAPE", Available: 3}}
	if got := service.Snapshot(); !reflect.DeepEqual(got, wantStock) {
		t.Fatalf("unexpected stock: %#v", got)
	}
}

func TestReserveOrderRejectsUnknownSKU(t *testing.T) {
	service := mustService(t, map[string]int{"TAPE": 4})
	_, err := service.ReserveOrder("order-2", []warehouse.Line{{SKU: "LABEL", Quantity: 1}})
	if !errors.Is(err, warehouse.ErrUnknownSKU) {
		t.Fatalf("expected ErrUnknownSKU, got %v", err)
	}
}

func TestReserveOrderRejectsInsufficientStock(t *testing.T) {
	service := mustService(t, map[string]int{"TAPE": 1})
	_, err := service.ReserveOrder("order-3", []warehouse.Line{{SKU: "TAPE", Quantity: 2}})
	if !errors.Is(err, warehouse.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

func TestReserveOrderRejectsInvalidLine(t *testing.T) {
	service := mustService(t, map[string]int{"TAPE": 4})
	_, err := service.ReserveOrder("order-4", []warehouse.Line{{SKU: "TAPE", Quantity: 0}})
	if !errors.Is(err, warehouse.ErrInvalidOrder) {
		t.Fatalf("expected ErrInvalidOrder, got %v", err)
	}
}

func TestReserveOrderRejectsCommittedOrderID(t *testing.T) {
	service := mustService(t, map[string]int{"TAPE": 4})
	if _, err := service.ReserveOrder("order-5", []warehouse.Line{{SKU: "TAPE", Quantity: 1}}); err != nil {
		t.Fatalf("first reservation failed: %v", err)
	}
	if _, err := service.ReserveOrder("order-5", []warehouse.Line{{SKU: "TAPE", Quantity: 1}}); !errors.Is(err, warehouse.ErrOrderExists) {
		t.Fatalf("expected ErrOrderExists, got %v", err)
	}
}

func TestNewServiceRejectsNormalizedDuplicateSKU(t *testing.T) {
	_, err := warehouse.NewService(map[string]int{"tape": 2, " TAPE ": 3})
	if !errors.Is(err, warehouse.ErrInvalidInventory) {
		t.Fatalf("expected ErrInvalidInventory, got %v", err)
	}
}

func mustService(t *testing.T, stock map[string]int) *warehouse.Service {
	t.Helper()
	service, err := warehouse.NewService(stock)
	if err != nil {
		t.Fatalf("NewService returned an error: %v", err)
	}
	return service
}
