package warehouse_test

import (
	"errors"
	"math"
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

func TestReserveOrderRejectsOverflowingQuantity(t *testing.T) {
	service := mustService(t, map[string]int{"BOX": math.MaxInt})
	wantStock := []warehouse.StockLevel{{SKU: "BOX", Available: math.MaxInt}}

	_, err := service.ReserveOrder("order-overflow", []warehouse.Line{
		{SKU: "BOX", Quantity: math.MaxInt},
		{SKU: "BOX", Quantity: 1},
	})
	if !errors.Is(err, warehouse.ErrInvalidOrder) {
		t.Fatalf("expected ErrInvalidOrder, got %v", err)
	}

	if got := service.Snapshot(); !reflect.DeepEqual(got, wantStock) {
		t.Fatalf("stock changed after failed reservation: %#v", got)
	}
	if _, exists := service.Reservation("order-overflow"); exists {
		t.Fatalf("order recorded after failed reservation")
	}
}

func TestReserveOrderRejectsOverflowAcrossManyLines(t *testing.T) {
	service := mustService(t, map[string]int{"BOX": math.MaxInt})
	half := math.MaxInt / 2
	rest := math.MaxInt - half
	_, err := service.ReserveOrder("order-many", []warehouse.Line{
		{SKU: "BOX", Quantity: half},
		{SKU: "BOX", Quantity: rest},
		{SKU: "BOX", Quantity: 1},
	})
	if !errors.Is(err, warehouse.ErrInvalidOrder) {
		t.Fatalf("expected ErrInvalidOrder, got %v", err)
	}
	if got := service.Snapshot(); !reflect.DeepEqual(got, []warehouse.StockLevel{{SKU: "BOX", Available: math.MaxInt}}) {
		t.Fatalf("stock changed after failed reservation: %#v", got)
	}
}

func TestReserveOrderAcceptsExactMaxIntSplit(t *testing.T) {
	service := mustService(t, map[string]int{"BOX": math.MaxInt})
	half := math.MaxInt / 2
	rest := math.MaxInt - half
	reservation, err := service.ReserveOrder("order-split", []warehouse.Line{
		{SKU: "BOX", Quantity: half},
		{SKU: "BOX", Quantity: rest},
	})
	if err != nil {
		t.Fatalf("ReserveOrder returned an error: %v", err)
	}
	wantLines := []warehouse.Line{{SKU: "BOX", Quantity: math.MaxInt}}
	if !reflect.DeepEqual(reservation.Lines, wantLines) {
		t.Fatalf("unexpected reservation lines: %#v", reservation.Lines)
	}
	if got := service.Snapshot(); !reflect.DeepEqual(got, []warehouse.StockLevel{{SKU: "BOX", Available: 0}}) {
		t.Fatalf("unexpected stock: %#v", got)
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
