package mtgbulk

import "sort"

type PossessionMatrix struct {
	SellerCards map[string]map[string]int
	CardSellers map[string]map[string]int
}

func NewPossessionMatrix() *PossessionMatrix {
	return &PossessionMatrix{
		SellerCards: make(map[string]map[string]int),
		CardSellers: make(map[string]map[string]int),
	}
}

func (m *PossessionMatrix) AddCard(seller, card string, price int) {
	c, ok := m.SellerCards[seller]
	if !ok {
		m.SellerCards[seller] = make(map[string]int)
		c = m.SellerCards[seller]
	}
	if c[card] == 0 || price < c[card] {
		c[card] = price
	}

	s, ok := m.CardSellers[card]
	if !ok {
		m.CardSellers[card] = make(map[string]int)
		s = m.CardSellers[card]
	}
	if s[seller] == 0 || price < s[seller] {
		s[seller] = price
	}
}

type PossessionTable struct {
	Sellers, Cards                     []string
	Prices                             [][]int
	SellerCardsTotal, SellerPriceTotal []int
	CardSellersTotal                   []int
}

func newPossessionTable(sellersN, cardsN int) *PossessionTable {
	t := &PossessionTable{}
	t.Sellers = make([]string, 0, sellersN)
	t.Cards = make([]string, 0, cardsN)
	t.Prices = make([][]int, 0, cardsN)
	for i := 0; i < cardsN; i++ {
		t.Prices = append(t.Prices, make([]int, sellersN))
	}
	t.SellerCardsTotal = make([]int, sellersN)
	t.SellerPriceTotal = make([]int, sellersN)
	t.CardSellersTotal = make([]int, cardsN)

	return t
}

func (m *PossessionMatrix) ToTable() *PossessionTable {
	cardsN := len(m.CardSellers)
	sellersN := len(m.SellerCards)
	t := newPossessionTable(sellersN, cardsN)

	for seller := range m.SellerCards {
		t.Sellers = append(t.Sellers, seller)
	}
	sort.Slice(t.Sellers, func(i, j int) bool {
		return len(m.SellerCards[t.Sellers[i]]) >= len(m.SellerCards[t.Sellers[j]])
	})

	for card := range m.CardSellers {
		t.Cards = append(t.Cards, card)
	}
	sort.Strings(t.Cards)

	for x, seller := range t.Sellers {
		for y, card := range t.Cards {
			p, ok := m.SellerCards[seller][card]
			t.Prices[y][x] = p
			t.SellerPriceTotal[x] = t.SellerPriceTotal[x] + p
			if ok {
				t.SellerCardsTotal[x] = t.SellerCardsTotal[x] + 1
				t.CardSellersTotal[y] = t.CardSellersTotal[y] + 1
			}
		}
	}
	return t
}
