package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/ilyalavrinov/mtgbulkbuy/pkg/mtgbulk"
	"go.uber.org/zap"
)

type handler struct {
	loggerRaw *zap.Logger
	logger    *zap.SugaredLogger
}

func newHandler() *handler {
	h := &handler{}
	var err error

	cfg := zap.NewDevelopmentConfig()
	cfg.Development = false
	h.loggerRaw, err = cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("logger init failed: %s", err))
	}
	h.logger = h.loggerRaw.Sugar()
	return h
}

func (h *handler) bulkHandler(resp http.ResponseWriter, req *http.Request) {
	defer h.logger.Sync()
	body := req.Body
	defer body.Close()

	contentType := req.Header["Content-Type"]
	if len(contentType) != 1 || contentType[0] != "text/plain" {
		resp.WriteHeader(http.StatusBadRequest)
		io.WriteString(resp, "Content-Type: text/plain is expected\n")
		return
	}

	result, err := h.handleText(body)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorw("Handle Text error",
			"err", err)
		return
	}

	resBody, err := json.Marshal(result.Cards)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorw("Cannot write json with text search result",
			"err", err)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write(resBody)
}

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

func (h *handler) handleText(r io.Reader) (*mtgbulk.NamesResult, error) {
	cards := mtgbulk.NewNamesRequest()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		h.logger.Debugw("New line read from body",
			"line", line)
		line = strings.Trim(line, " ")
		if len(line) == 0 {
			continue
		}
		name, quantity, err := parseLine(line)
		if err != nil {
			h.logger.Warnw("could not parse line",
				"err", err,
				"line", line)
			return nil, err
		}

		h.logger.Debugw("Parsed line",
			"line", line,
			"cardname", name,
			"quantity", quantity)

		if _, found := cards.Cards[name]; found {
			h.logger.Warnw("Duplicated card",
				"name", name)
			return nil, fmt.Errorf("Card with name %q is duplicated in the list", name)
		}

		if quantity <= 0 {
			h.logger.Warnw("Illegal requested quantity",
				"name", name,
				"quantity", quantity)
			return nil, fmt.Errorf("Illegal quantity for card %q has been requested: %d", name, quantity)
		}

		cards.Cards[name] = quantity
	}
	if err := scanner.Err(); err != nil {
		h.logger.Warnw("Error reading body",
			"err", err)
		return nil, err
	}

	if len(cards.Cards) == 0 {
		h.logger.Warnw("Empty card list")
		return nil, fmt.Errorf("Empty card list")
	}

	result, err := mtgbulk.ProcessByNames(cards)
	if err != nil {
		h.logger.Warnw("Could not process request",
			"err", err)
		return nil, err
	}

	return result, nil
}
