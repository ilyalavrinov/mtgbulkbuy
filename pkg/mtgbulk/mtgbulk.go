package mtgbulk

import (
	"encoding/json"
	"fmt"
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

func (c *CardResult) merge(other CardResult) {
	if c.Available || other.Available {
		c.Available = true
	}
	c.Prices = append(c.Prices, other.Prices...)
}

func (c *CardResult) sortByPrice() {
	sort.Slice(c.Prices, func(i, j int) bool {
		return c.Prices[i].Price < c.Prices[j].Price
	})
}

type NamesResult struct {
	RawCards map[string]CardResult

	MinPricesRule map[string][]CardPrice
}

func ProcessByNames(cards NamesRequest) (*NamesResult, error) {
	logger.Debugw("Incoming ProcessByNames request",
		"count", len(cards.Cards))

	result := &NamesResult{
		RawCards: make(map[string]CardResult, len(cards.Cards)),
	}
	for name := range cards.Cards {
		cardRes := newCardResult()
		cardRes.merge(searchMtgSale(name))
		cardRes.merge(searchMtgTrade(name))
		cardRes.sortByPrice()
		result.RawCards[name] = cardRes
	}

	greedyMinPrices, err := calcGreedyMinPrices(cards, result.RawCards)
	if err != nil {
		logger.Errorw("could not calculate greedy min prices",
			"err", err)
		return result, err
	}
	result.MinPricesRule = greedyMinPrices

	return result, nil
}

func calcGreedyMinPrices(req NamesRequest, cards map[string]CardResult) (map[string][]CardPrice, error) {
	result := make(map[string][]CardPrice, len(req.Cards))

	for name, reqCount := range req.Cards {
		cardData, found := cards[name]
		if !found {
			return nil, fmt.Errorf("Card %q not found", name)
		}

		if !cardData.Available {
			return nil, fmt.Errorf("Card %q is not available", name)
		}

		cardsFound := 0
		for _, p := range cardData.Prices {
			if cardsFound >= reqCount {
				break
			}
			toAdd := p
			if toAdd.Quantity > reqCount-cardsFound {
				toAdd.Quantity = reqCount - cardsFound
			}
			result[name] = append(result[name], toAdd)
			cardsFound += toAdd.Quantity
			logger.Debugw("greedy min price add result",
				"name", name,
				"qty", toAdd.Quantity,
				"price", toAdd.Price,
				"totalFound", cardsFound,
				"reqCount", reqCount)
		}
	}

	return result, nil
}
