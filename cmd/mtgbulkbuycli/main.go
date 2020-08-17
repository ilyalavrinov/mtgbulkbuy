package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/ilyalavrinov/mtgbulkbuy/internal/texthandler"
	"github.com/ilyalavrinov/mtgbulkbuy/pkg/mtgbulk"
	"github.com/jedib0t/go-pretty/table"
	"github.com/tealeg/xlsx"
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

	res := *filename + ".matrix.out"
	os.Remove(res)
	f, err = os.Create(res)
	defer f.Close()
	if err != nil {
		fmt.Printf("Could not create file possession matrix, aborting")
		return
	}
	t := mtgbulk.NewPossessionTable(result.MinPricesMatrix)
	t.ToTextTable(f)
	if err != nil {
		fmt.Printf("Could not write possession matrix, aborting")
		return
	}

	xls := xlsx.NewFile()
	sh, err := xls.AddSheet("min_prices_all")
	if err != nil {
		fmt.Printf("Could not add min prices table xlsx sheet, aborting")
		return
	}
	err = t.ToXlsxSheet(sh)
	if err != nil {
		fmt.Printf("Could not fill min prices table xlsx sheet, aborting")
		return
	}

	xlsname := *filename + ".xlsx"
	os.Remove(xlsname)
	fxls, err := os.Create(xlsname)
	if err != nil {
		fmt.Printf("Could not write possession matrix xlsx file, aborting")
		return
	}
	err = xls.Write(fxls)
	if err != nil {
		fmt.Printf("Could not write possession matrix xlsx, aborting")
		return
	}
}
