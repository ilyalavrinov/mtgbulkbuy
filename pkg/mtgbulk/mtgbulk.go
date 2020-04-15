package mtgbulk

import (
	"encoding/json"
	"sort"
)

type NamesRequest struct {
	Cards map[string]int
}

func NewNamesRequest() NamesRequest {
	return NamesRequest{
		Cards: make(map[string]int),
	}
}

type PlatformType int

const (
	MtgSale  PlatformType = iota
	MtgTrade PlatformType = iota
)

type CurrencyType int

const (
	RUR CurrencyType = iota
	USD CurrencyType = iota
)

func (c CurrencyType) String() string {
	res := ""
	switch c {
	case RUR:
		res = "₽"
	case USD:
		res = "$"
	}
	return res
}

func (c CurrencyType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

type CardPrice struct {
	Price    float32
	Foil     bool
	Currency CurrencyType
	Quantity int

	Platform PlatformType
	Trader   string
	URL      string
}

type CardResult struct {
	Available bool
	Prices    []CardPrice
}

func newCardResult() CardResult {
	return CardResult{
		Available: false,
		Prices:    make([]CardPrice, 0),
	}
}

func (c *CardResult) sortByPrice() {
	sort.Slice(c.Prices, func(i, j int) bool {
		return c.Prices[i].Price < c.Prices[j].Price
	})
}

type NamesResult struct {
	Cards map[string]CardResult
}

func ProcessByNames(cards NamesRequest) (*NamesResult, error) {
	logger.Debugw("Incoming ProcessByNames request",
		"count", len(cards.Cards))

	result := &NamesResult{
		Cards: make(map[string]CardResult, len(cards.Cards)),
	}
	for name := range cards.Cards {
		res := searchMtgSale(name)
		res.sortByPrice()
		result.Cards[name] = res
	}

	return result, nil
}
