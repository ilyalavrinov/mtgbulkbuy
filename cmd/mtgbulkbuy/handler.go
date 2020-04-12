package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"

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
	h.loggerRaw, err = zap.NewDevelopment()
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
		io.WriteString(resp, "Content-Type: text/plain is expected")
		return
	}

	err := h.handleText(body)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		io.WriteString(resp, fmt.Sprintf("Incoming text cannot be handled"))
	}
}

func (h *handler) handleText(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		h.logger.Debugw("New line read from body",
			"line", line)
	}
	if err := scanner.Err(); err != nil {
		h.logger.Errorw("Error reading body",
			"err", err)
		return err
	}

	// TODO: parse lines
	cards := mtgbulk.NamesRequest{}

	_, err := mtgbulk.ProcessByNames(cards)
	if err != nil {
		fmt.Printf("Could not process request for %d cards; error: %s\n", len(cards.Cards), err)
	}

	return nil
}
