package main

import (
	"log"
	"net/http"

	"github.com/dashhive/p2pwebproxy"
)

func main() {
	bindAddr := ":8080"
	log.Printf("Listening on %s", bindAddr)
	http.HandleFunc("OPTIONS /ws", p2pwebproxy.AddCORSHandler)
	http.HandleFunc("GET /ws", p2pwebproxy.Handler)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}
