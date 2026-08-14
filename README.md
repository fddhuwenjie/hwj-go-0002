# stockreserve

`stockreserve` is a small warehouse reservation service and command-line tool. It validates a complete order before changing stock, aggregates repeated SKUs inside an order, and reports domain errors without requiring an external database.

Inventory JSON uses a `stock` object and an order contains an `order_id` plus line items:

```json
{"stock":{"BOX-SMALL":12,"TAPE":8}}
```

```json
{"order_id":"order-104","lines":[{"sku":"box-small","quantity":2},{"sku":"tape","quantity":1}]}
```

Run the project with:

```bash
go test ./...
go run ./cmd/stockreserve inventory.json order.json
```
