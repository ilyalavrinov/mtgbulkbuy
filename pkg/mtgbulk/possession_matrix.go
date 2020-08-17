package mtgbulk

import (
	"fmt"
	"io"
	"sort"

	"github.com/jedib0t/go-pretty/table"
	"github.com/tealeg/xlsx"
)

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

func NewPossessionTable(m *PossessionMatrix) *PossessionTable {
	t := &PossessionTable{}

	cardsN := len(m.CardSellers)
	sellersN := len(m.SellerCards)

	t.Sellers = make([]string, 0, sellersN)
	t.Cards = make([]string, 0, cardsN)
	t.Prices = make([][]int, 0, cardsN)
	for i := 0; i < cardsN; i++ {
		t.Prices = append(t.Prices, make([]int, sellersN))
	}
	t.SellerCardsTotal = make([]int, sellersN)
	t.SellerPriceTotal = make([]int, sellersN)
	t.CardSellersTotal = make([]int, cardsN)

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

func (t *PossessionTable) ToTextTable(out io.Writer) error {
	tOut := table.NewWriter()

	tOut.SetOutputMirror(out)
	header := make(table.Row, 0, len(t.Sellers)+2)
	header = append(header, "CARD\\\\SELLER")
	for _, s := range t.Sellers {
		header = append(header, s)
	}
	header = append(header, "TOTAL SELLERS")
	tOut.AppendHeader(header)

	rows := make([]table.Row, 0, len(t.Cards))
	for ci, pr := range t.Prices {
		row := make(table.Row, 0, len(t.Sellers)+1)
		row = append(row, t.Cards[ci])
		for _, p := range pr {
			row = append(row, p)
		}
		row = append(row, t.CardSellersTotal[ci])
		rows = append(rows, row)
	}
	tOut.AppendRows(rows)

	f1 := make(table.Row, 0, len(t.Sellers)+2)
	f1 = append(f1, "Total price")
	for _, p := range t.SellerPriceTotal {
		f1 = append(f1, p)
	}
	tOut.AppendFooter(f1)

	f2 := make(table.Row, 0, len(t.Sellers)+2)
	f2 = append(f2, "Total cards")
	for _, c := range t.SellerCardsTotal {
		f2 = append(f2, c)
	}
	tOut.AppendFooter(f2)
	tOut.Render()

	return nil
}

func (t *PossessionTable) ToXlsxSheet(out *xlsx.Sheet, minPrices map[string]int) error {
	xOffset := 1
	yOffset := 0
	for x, seller := range t.Sellers {
		c := out.Cell(yOffset+0, xOffset+x)
		c.SetString(seller)
	}

	xOffset = 0
	yOffset = 1
	for y, card := range t.Cards {
		c := out.Cell(yOffset+y, xOffset+0)
		c.SetString(card)
	}

	xOffset = 1
	yOffset = 1
	minPriceStyle := xlsx.NewStyle()
	minPriceStyle.ApplyFill = true
	minPriceStyle.Fill.PatternType = xlsx.Solid_Cell_Fill
	minPriceStyle.Fill.FgColor = xlsx.RGB_Light_Green
	for y, row := range t.Prices {
		for x, p := range row {
			c := out.Cell(yOffset+y, xOffset+x)
			c.SetInt(p)
			pMin := minPrices[t.Cards[y]]
			if pMin != 0 && p == pMin {
				c.SetStyle(minPriceStyle)
			}
		}
	}

	yOffset += len(t.Cards) + 1
	c := out.Cell(yOffset+0, 0)
	c.SetString("TOTAL PRICE")
	for x := range t.SellerPriceTotal {
		c := out.Cell(0+yOffset, x+xOffset)
		// c.SetInt(val)
		formula := fmt.Sprintf("=SUM(%s2:%s%d)", xlsx.ColIndexToLetters(xOffset+x), xlsx.ColIndexToLetters(xOffset+x), 1+len(t.Cards))
		c.SetStringFormula(formula)
	}

	yOffset += 1
	c = out.Cell(yOffset+0, 0)
	c.SetString("TOTAL CARDS")
	for x, totals := range t.SellerCardsTotal {
		c := out.Cell(0+yOffset, x+xOffset)
		c.SetInt(totals)
	}

	return nil
}
