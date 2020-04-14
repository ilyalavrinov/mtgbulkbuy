package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	h := newHandler()
	defer h.loggerRaw.Sync()

	h.logger.Debug("Registering handlers")
	router.HandleFunc("/bulk", h.bulkHandler)
	h.logger.Debug("Registration finished")

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	h.logger.Debugw("start listening",
		"addr", srv.Addr)
	err := srv.ListenAndServe()
	h.logger.Fatalw("listen failed",
		"err", err)
	h.logger.Debug("listen finished. Exiting")
}
