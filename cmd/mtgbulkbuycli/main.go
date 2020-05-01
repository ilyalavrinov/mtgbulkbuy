package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyalavrinov/mtgbulkbuy/internal/texthandler"
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

	for name, cards := range result.RawCards {
		fmt.Printf("%s ==> total found %d\n", name, len(cards.Prices))
	}

	if len(result.MinPricesRule) > 0 {
		fmt.Println("Min price rule:")
		for name, prices := range result.MinPricesRule {
			for _, p := range prices {
				fmt.Printf("%s ==> qty: %d; price: %.0f; seller: %s\n", name, p.Quantity, p.Price, p.Trader)
			}
		}
	}
}
