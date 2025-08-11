package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func handlerChunkedResponse(w http.ResponseWriter, r *http.Request) {
	// Content-Length を設定しないことで自動的に chunked になる
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	c := http.NewResponseController(w)
	for i := 1; i <= 10; i++ {
		if _, err := fmt.Fprintf(w, "Chunk #%d\n", i); err != nil {
			log.Println("write error:", err)
			return
		}
		if err := c.Flush(); err != nil {
			log.Println("flush error:", err)
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	http.HandleFunc("/chunked", handlerChunkedResponse)

	addr := ":18888"
	log.Println("listening on", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
