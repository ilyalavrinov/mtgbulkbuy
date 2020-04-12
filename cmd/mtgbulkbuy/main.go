package main

import (
	"fmt"

	"github.com/ilyalavrinov/mtgbulkbuy/pkg/mtgbulk"
)

func main() {
	tmpCards := mtgbulk.NamesRequest{}

	_, err := mtgbulk.ProcessByNames(tmpCards)
	if err != nil {
		fmt.Printf("Could not process request for %d cards; error: %s\n", len(tmpCards.Cards), err)
	}
}
