package orderbook

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type Order struct {
	Size      float64
	Bid       bool
	Limit     *Limit
	Timestamp int64
	ID        int64
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

func (order *Order) String() string {
	return fmt.Sprintf("Order Size: %.2f", order.Size)
}

func (order *Order) IsFilled() bool {
	return order.Size == 0.0
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size:      size,
		Bid:       bid,
		Timestamp: time.Now().UnixNano(),
		ID:        int64(rand.Intn(10000000000)),
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

func (limit *Limit) Fill(order *Order) []Match {
	matches := []Match{}
	ordersToDelete := []*Order{}

	for _, o := range limit.Orders {
		match := limit.fillOrder(o, order)
		matches = append(matches, match)
		limit.Volume -= match.SizeFilled
		if o.IsFilled() {
			ordersToDelete = append(ordersToDelete, o)
		}
		if order.IsFilled() {
			break
		}
	}

	for _, o := range ordersToDelete {
		limit.DeleteOrder(o)
	}

	return matches
}

func (limit *Limit) fillOrder(a, b *Order) Match {
	var bid *Order
	var ask *Order
	var sizeFilled float64

	if a.Bid {
		bid = a
		ask = b
	} else {
		bid = b
		ask = a
	}

	if a.Size >= b.Size {
		a.Size -= b.Size
		sizeFilled = b.Size
		b.Size = 0.0
	} else {
		b.Size -= a.Size
		sizeFilled = a.Size
		a.Size = 0.0
	}

	return Match{
		Bid:        bid,
		Ask:        ask,
		SizeFilled: sizeFilled,
		Price:      limit.Price,
	}
}

type Orderbook struct {
	asks []*Limit
	bids []*Limit

	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func NewOrderBook() *Orderbook {
	return &Orderbook{
		asks:      []*Limit{},
		bids:      []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

func (orderbook *Orderbook) PlaceMarketOrder(order *Order) []Match {
	matches := []Match{}

	if order.Bid {
		if order.Size > orderbook.AskTotalVolume() {
			panic(fmt.Errorf("Not enough volume [size: %.2f] for market order [size: %.2f]", orderbook.AskTotalVolume(), order.Size))
		}

		for _, limit := range orderbook.Asks() {
			limitMatches := limit.Fill(order)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				orderbook.clearLimit(true, limit)
			}
		}
	} else {
		if order.Size > orderbook.BidTotalVolume() {
			panic(fmt.Errorf("Not enough volume [size: %.2f] for market order [size: %.2f]", orderbook.BidTotalVolume(), order.Size))
		}
		for _, limit := range orderbook.Bids() {
			limitMatches := limit.Fill(order)
			matches = append(matches, limitMatches...)

			if len(limit.Orders) == 0 {
				orderbook.clearLimit(true, limit)
			}
		}
	}

	return matches
}

func (orderbook *Orderbook) PlaceLimitOrder(price float64, order *Order) {
	var limit *Limit
	if order.Bid {
		limit = orderbook.BidLimits[price]
	} else {
		limit = orderbook.AskLimits[price]
	}

	if limit == nil {
		limit = NewLimit(price)
		if order.Bid {
			orderbook.bids = append(orderbook.bids, limit)
			orderbook.BidLimits[price] = limit
		} else {
			orderbook.asks = append(orderbook.asks, limit)
			orderbook.AskLimits[price] = limit
		}
	}
	limit.AddOrder(order)
}

func (orderbook *Orderbook) clearLimit(bid bool, limit *Limit) {
	if bid {
		delete(orderbook.BidLimits, limit.Price)
		for i := 0; i < len(orderbook.bids); i++ {
			if orderbook.bids[i] == limit {
				orderbook.bids[i] = orderbook.bids[len(orderbook.bids)-1]
				orderbook.bids = orderbook.bids[:len(orderbook.bids)-1]
			}
		}
	} else {
		delete(orderbook.AskLimits, limit.Price)
		for i := 0; i < len(orderbook.bids); i++ {
			if orderbook.asks[i] == limit {
				orderbook.asks[i] = orderbook.asks[len(orderbook.asks)-1]
				orderbook.asks = orderbook.asks[:len(orderbook.asks)-1]
			}
		}
	}
}

func (orderbook *Orderbook) CancelOrder(order *Order) {
	limit := order.Limit
	limit.DeleteOrder(order)
}

func (orderbook *Orderbook) BidTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(orderbook.bids); i++ {
		totalVolume += orderbook.bids[i].Volume
	}

	return totalVolume
}

func (orderbook *Orderbook) AskTotalVolume() float64 {
	totalVolume := 0.0

	for i := 0; i < len(orderbook.asks); i++ {
		totalVolume += orderbook.asks[i].Volume
	}

	return totalVolume
}

func (orderbook *Orderbook) Asks() []*Limit {
	sort.Sort(BestAskPrice{orderbook.asks})
	return orderbook.asks
}

func (orderbook *Orderbook) Bids() []*Limit {
	sort.Sort(BestBidPrice{orderbook.bids})
	return orderbook.bids
}
