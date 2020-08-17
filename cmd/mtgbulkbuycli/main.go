package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/ilyalavrinov/mtgbulkbuy/internal/texthandler"
	"github.com/jedib0t/go-pretty/table"
)

const (
	filenameArg   = "from"
	filenameUsage = "file with list of cards to be processed"
)

var filename = flag.String(filenameArg, "", filenameUsage)

func main() {
	flag.Parse()

	if *filename == "" {
		fmt.Printf("%q arg is mandatory\n", filenameArg)
		os.Exit(1)
	}

	f, err := os.Open(*filename)
	if err != nil {
		fmt.Printf("could not open file %q; error: %s", *filename, err)
		os.Exit(1)
	}

	result, err := texthandler.HandleText(f)
	if err != nil {
		fmt.Printf("could not get result; error: %s", err)
		os.Exit(1)
	}

	for name, cards := range result.AllSortedCards {
		fmt.Printf("%s ==> total found %d\n", name, len(cards.Prices))
	}

	if len(result.MinPricesNoDelivery) > 0 {
		fmt.Println("Min price rule:")
		var total float32
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Cardname", "Qty", "Price", "Seller"})
		rows := make([]table.Row, 0)
		for name, prices := range result.MinPricesNoDelivery {
			for _, p := range prices {
				rows = append(rows, table.Row{name, p.Quantity, p.Price, p.Trader})
				total += p.Price
			}
		}
		sort.Slice(rows, func(i, j int) bool {
			return rows[i][0].(string) < rows[j][0].(string)
		})
		t.AppendRows(rows)
		t.AppendFooter(table.Row{"", "", "Total", total})
		t.Render()
	}

	{
		res := *filename + ".matrix.out"
		os.Remove(res)
		f, err := os.Create(res)
		if err != nil {
			fmt.Printf("Could not write possession matrix, aborting")
			return
		}
		t := table.NewWriter()
		t.SetOutputMirror(f)
		posT := result.MinPricesMatrix.ToTable()
		header := make(table.Row, 0, len(posT.Sellers)+2)
		header = append(header, "CARD\\\\SELLER")
		for _, s := range posT.Sellers {
			header = append(header, s)
		}
		header = append(header, "TOTAL SELLERS")
		t.AppendHeader(header)

		rows := make([]table.Row, 0, len(posT.Cards))
		for ci, pr := range posT.Prices {
			row := make(table.Row, 0, len(posT.Sellers)+1)
			row = append(row, posT.Cards[ci])
			for _, p := range pr {
				row = append(row, p)
			}
			row = append(row, posT.CardSellersTotal[ci])
			rows = append(rows, row)
		}
		t.AppendRows(rows)

		f1 := make(table.Row, 0, len(posT.Sellers)+2)
		f1 = append(f1, "Total price")
		for _, p := range posT.SellerPriceTotal {
			f1 = append(f1, p)
		}
		t.AppendFooter(f1)

		f2 := make(table.Row, 0, len(posT.Sellers)+2)
		f2 = append(f2, "Total cards")
		for _, c := range posT.SellerCardsTotal {
			f2 = append(f2, c)
		}
		t.AppendFooter(f2)
		t.Render()
	}
}
