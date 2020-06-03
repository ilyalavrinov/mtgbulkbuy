package mtgbulk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly"
)

func searchMtgTrade(cardname string) CardResult {
	cardname = strings.ToLower(cardname)
	result := newCardResult()
	addr := mtgTradeSearchURL(cardname)

	c := colly.NewCollector()
	c.OnHTML(".search-item", func(e *colly.HTMLElement) {
		if strings.ToLower(e.ChildText(".catalog-title")) != cardname {
			return
		}
		e.ForEach("table.search-card", func(i int, eTable *colly.HTMLElement) {
			trader := eTable.ChildText("tbody .trader-name a")

			eTable.ForEach("tbody tr", func(i int, eTR *colly.HTMLElement) {
				price, err := strconv.ParseFloat(eTR.ChildText(".catalog-rate-price"), 32)
				if err != nil {
					logger.Errorw("card price convert failed",
						"err", err)
					return
				}

				quantity, err := strconv.Atoi(eTR.ChildText(".sale-count"))
				if err != nil {
					logger.Errorw("card count convert failed",
						"err", err)
					return
				}

				foil := false
				if eTR.ChildAttr("img.foil", "src") != "" {
					foil = true
				}

				logger.Debugw("card",
					"row_index", i,
					"trader", trader,
					"price", price,
					"count", quantity,
					"foil", foil,
					"quality", eTR.ChildText(".js-card-quality-tooltip")) // TODO: to CardPrice

				result.Available = true
				result.Prices = append(result.Prices, CardPrice{
					Price:    float32(price),
					Foil:     foil,
					Currency: RUR,
					Quantity: quantity,
					Platform: MtgTrade,
					Trader:   trader,
					URL:      addr, // TODO: correct it! - it's just a search result, but we can get a direct link to a card at a seller
				})
			})
		})
	})

	err := c.Visit(addr)
	if err != nil {
		logger.Errorw("Unable to visit with scraper",
			"url", addr,
			"err", err)
	}
	return result
}

func mtgTradeSearchURL(cardname string) string {
	cardname = strings.ReplaceAll(cardname, " ", "+")
	return fmt.Sprintf("http://mtgtrade.net/search/?query=%s", cardname)
}
