package orderjson_test

import (
	"strings"
	"testing"

	"github.com/fddhuwenjie/hwj-go-0002/orderjson"
	"github.com/fddhuwenjie/hwj-go-0002/warehouse"
)

func TestDecodeInventoryAndReserve(t *testing.T) {
	stock, err := orderjson.DecodeInventory(strings.NewReader(`{"stock":{"BOX":5}}`))
	if err != nil {
		t.Fatalf("DecodeInventory returned an error: %v", err)
	}
	service, err := warehouse.NewService(stock)
	if err != nil {
		t.Fatalf("NewService returned an error: %v", err)
	}
	reservation, err := orderjson.Reserve(service, strings.NewReader(`{"order_id":"json-1","lines":[{"sku":"box","quantity":2}]}`))
	if err != nil {
		t.Fatalf("Reserve returned an error: %v", err)
	}
	if reservation.OrderID != "json-1" || len(reservation.Lines) != 1 || reservation.Lines[0].SKU != "BOX" {
		t.Fatalf("unexpected reservation: %#v", reservation)
	}
}

func TestDecodeOrderRejectsUnknownField(t *testing.T) {
	_, err := orderjson.DecodeOrder(strings.NewReader(`{"order_id":"json-2","lines":[],"priority":true}`))
	if err == nil {
		t.Fatal("DecodeOrder accepted an unknown field")
	}
}
