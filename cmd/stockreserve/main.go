package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fddhuwenjie/hwj-go-0002/orderjson"
	"github.com/fddhuwenjie/hwj-go-0002/warehouse"
)

type output struct {
	Reservation warehouse.Reservation  `json:"reservation"`
	Stock       []warehouse.StockLevel `json:"stock"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: stockreserve INVENTORY.json ORDER.json")
		os.Exit(2)
	}
	inventoryFile, err := os.Open(os.Args[1])
	if err != nil {
		fail(fmt.Errorf("open inventory: %w", err))
	}
	defer inventoryFile.Close()
	stock, err := orderjson.DecodeInventory(inventoryFile)
	if err != nil {
		fail(err)
	}
	service, err := warehouse.NewService(stock)
	if err != nil {
		fail(err)
	}
	orderFile, err := os.Open(os.Args[2])
	if err != nil {
		fail(fmt.Errorf("open order: %w", err))
	}
	defer orderFile.Close()
	reservation, err := orderjson.Reserve(service, orderFile)
	if err != nil {
		fail(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{Reservation: reservation, Stock: service.Snapshot()}); err != nil {
		fail(fmt.Errorf("encode result: %w", err))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
