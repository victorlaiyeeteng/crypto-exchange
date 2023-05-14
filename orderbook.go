package main

import (
	"fmt"
	"sort"
	"time"
)

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
}

type Orders []*Order

func (orders Orders) Len() int {
	return len(orders)
}
func (orders Orders) Swap(i, j int) {
	orders[i], orders[j] = orders[j], orders[i]
}
func (orders Orders) Less(i, j int) bool {
	return orders[i].Timestamp < orders[j].Timestamp
}

func (o *Order) String() string {
	return fmt.Sprintf("Order Size: %.2f", o.Size)
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
	}
}

type Match struct {
	Ask   *Order
	Bid   *Order
	Price float64
	// Manage the size left for the bid order to be completed
	SizeFilled float64
}

type Limit struct {
	Price  float64
	Orders Orders
	Volume float64
}

// To sort the limits
type Limits []*Limit
type BestAskPrice struct{ Limits }

func (ask BestAskPrice) Len() int {
	return len(ask.Limits)
}
func (ask BestAskPrice) Swap(i, j int) {
	ask.Limits[i], ask.Limits[j] = ask.Limits[j], ask.Limits[i]
}
func (ask BestAskPrice) Less(i, j int) bool {
	return ask.Limits[i].Price < ask.Limits[j].Price
}

type BestBidPrice struct{ Limits }

func (bid BestBidPrice) Len() int {
	return len(bid.Limits)
}
func (bid BestBidPrice) Swap(i, j int) {
	bid.Limits[i], bid.Limits[j] = bid.Limits[j], bid.Limits[i]
}
func (bid BestBidPrice) Less(i, j int) bool {
	return bid.Limits[i].Price > bid.Limits[j].Price
}

func (limit *Limit) String() string {
	return fmt.Sprintf("[Price: %.2f | Volume: %.2f]", limit.Price, limit.Volume)
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price:  price,
		Orders: []*Order{},
	}
}

func (limit *Limit) AddOrder(order *Order) {
	order.Limit = limit
	limit.Orders = append(limit.Orders, order)
	limit.Volume += order.Size
}

func (limit *Limit) DeleteOrder(order *Order) {
	for i := 0; i < len(limit.Orders); i++ {
		if limit.Orders[i] == order {
			limit.Orders = append(limit.Orders[:i], limit.Orders[i+1:]...)
		}
	}
	limit.Volume -= order.Size
	order.Limit = nil
	sort.Sort(limit.Orders)
}

type Orderbook struct {
	Asks []*Limit
	Bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func NewOrderBook() *Orderbook {
	return &Orderbook{
		Asks:      []*Limit{},
		Bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

func (orderbook *Orderbook) PlaceOrder(price float64, order *Order) []Match {
	// Attempt to match order to those in the orderbook

	// Add the remainder of the order back to the orderbook
	if order.Size > 0.0 {
		orderbook.add(price, order)
	}

	return []Match{}
}

func (orderbook *Orderbook) add(price float64, order *Order) {
	var limit *Limit
	if order.Bid {
		limit = orderbook.BidLimits[price]
	} else {
		limit = orderbook.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		if order.Bid {
			orderbook.Bids = append(orderbook.Bids, limit)
			orderbook.BidLimits[price] = limit
		} else {
			orderbook.Asks = append(orderbook.Asks, limit)
			orderbook.AskLimits[price] = limit
		}
	}
	limit.AddOrder(order)
}
