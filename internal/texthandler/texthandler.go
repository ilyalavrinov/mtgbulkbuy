package texthandler

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/ilyalavrinov/mtgbulkbuy/pkg/mtgbulk"
)

var quantityRe *regexp.Regexp = regexp.MustCompile("^(\\d+)x?\\s*(.*)$")

func parseLine(line string) (string, int, error) {
	quantity := 1
	cardname := line
	var err error
	if quantityRe.MatchString(line) {
		matches := quantityRe.FindAllStringSubmatch(line, -1)

		quantity, err = strconv.Atoi(matches[0][1])
		if err != nil {
			return "", 0, err
		}
		cardname = matches[0][2]
		if len(cardname) == 0 {
			return "", 0, fmt.Errorf("empty cardname")
		}
	}

	return cardname, quantity, nil
}

func HandleText(r io.Reader) (*mtgbulk.NamesResult, error) {
	cards := mtgbulk.NewNamesRequest()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Debugw("New line read from body",
			"line", line)
		line = strings.Trim(line, " ")
		if len(line) == 0 {
			continue
		}
		name, quantity, err := parseLine(line)
		if err != nil {
			logger.Warnw("could not parse line",
				"err", err,
				"line", line)
			return nil, err
		}

		logger.Debugw("Parsed line",
			"line", line,
			"cardname", name,
			"quantity", quantity)

		if _, found := cards.Cards[name]; found {
			logger.Warnw("Duplicated card",
				"name", name)
			return nil, fmt.Errorf("Card with name %q is duplicated in the list", name)
		}

		if quantity <= 0 {
			logger.Warnw("Illegal requested quantity",
				"name", name,
				"quantity", quantity)
			return nil, fmt.Errorf("Illegal quantity for card %q has been requested: %d", name, quantity)
		}

		cards.Cards[name] = quantity
	}
	if err := scanner.Err(); err != nil {
		logger.Warnw("Error reading body",
			"err", err)
		return nil, err
	}

	if len(cards.Cards) == 0 {
		logger.Warnw("Empty card list")
		return nil, fmt.Errorf("Empty card list")
	}

	result, err := mtgbulk.ProcessByNames(cards)
	if err != nil {
		logger.Warnw("Could not process request",
			"err", err)
		return nil, err
	}

	return result, nil
}
