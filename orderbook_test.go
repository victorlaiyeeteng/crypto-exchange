package main

import (
	"fmt"
	"reflect"
	"testing"
)

func assert(t *testing.T, a, b any) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("%+v != %+v", a, b)
	}
}

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

func TestPlaceLimitOrder(t *testing.T) {
	orderbook := NewOrderBook()

	sellOrderA := NewOrder(false, 10)
	sellOrderB := NewOrder(false, 5)
	orderbook.PlaceLimitOrder(10_000, sellOrderA)
	orderbook.PlaceLimitOrder(9_000, sellOrderB)

	assert(t, len(orderbook.asks), 2)
}

func TestPlaceMarketOrder(t *testing.T) {
	orderbook := NewOrderBook()

	sellOrder := NewOrder(false, 20)
	orderbook.PlaceLimitOrder(10_000, sellOrder)

	buyOrder := NewOrder(true, 10)
	matches := orderbook.PlaceMarketOrder(buyOrder)

	assert(t, len(matches), 1)
	assert(t, len(orderbook.asks), 1)
	assert(t, orderbook.AskTotalVolume(), 10.0)
	assert(t, matches[0].Ask, sellOrder)
	assert(t, matches[0].Bid, buyOrder)
	assert(t, matches[0].SizeFilled, 10.0)
	assert(t, matches[0].Price, 10_000.0)
	assert(t, buyOrder.IsFilled(), true)

	fmt.Printf("%+v", matches)
}

func TestPlaceMarketOrderMultiFill(t *testing.T) {
	orderbook := NewOrderBook()

	buyOrderA := NewOrder(true, 5)
	buyOrderB := NewOrder(true, 8)
	buyOrderC := NewOrder(true, 10)
	buyOrderD := NewOrder(true, 1)

	orderbook.PlaceLimitOrder(5_000, buyOrderD)
	orderbook.PlaceLimitOrder(5_000, buyOrderC)
	orderbook.PlaceLimitOrder(9_000, buyOrderB)
	orderbook.PlaceLimitOrder(10_000, buyOrderA)

	assert(t, orderbook.BidTotalVolume(), 24.00)

	sellOrder := NewOrder(false, 20)
	matches := orderbook.PlaceMarketOrder(sellOrder)

	assert(t, orderbook.BidTotalVolume(), 4.0)
	assert(t, len(matches), 4)
	assert(t, len(orderbook.bids), 1)

	fmt.Printf("%+v", matches)

}

func TestCancelOrder(t *testing.T) {
	orderbook := NewOrderBook()

	buyOrder := NewOrder(true, 10)

	orderbook.PlaceLimitOrder(10_000, buyOrder)

	assert(t, orderbook.BidTotalVolume(), 10.0)

	orderbook.cancelOrder(buyOrder)

	assert(t, orderbook.BidTotalVolume(), 0.0)
}
