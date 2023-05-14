package main

import (
	"fmt"
	"testing"
)

func TestLimit(t *testing.T) {
	limit := NewLimit(10_000)
	bidOrder1 := NewOrder(true, 5)
	bidOrder2 := NewOrder(true, 10)
	bidOrder3 := NewOrder(true, 15)
	limit.AddOrder(bidOrder1)
	limit.AddOrder(bidOrder2)
	limit.AddOrder(bidOrder3)
	limit.DeleteOrder(bidOrder2)
	fmt.Println(limit)
}

func TestOrderBook(t *testing.T) {
	orderbook := NewOrderBook()

	bidOrder1 := NewOrder(true, 10)
	bidOrder2 := NewOrder(true, 200)

	orderbook.PlaceOrder(15_000, bidOrder1)
	orderbook.PlaceOrder(18_000, bidOrder2)

	fmt.Printf("%+v", orderbook.Bids)
}
