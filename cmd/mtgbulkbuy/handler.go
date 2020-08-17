package main

import (
	"encoding/json"
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

	result, err := mtgbulk.ProcessText(body)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorw("Handle Text error",
			"err", err)
		return
	}

	resBody, err := json.Marshal(result.MinPricesNoDelivery)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		h.logger.Errorw("Cannot write json with text search result",
			"err", err)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write(resBody)
}
