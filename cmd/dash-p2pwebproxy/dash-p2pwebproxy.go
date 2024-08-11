package main

import (
	"log"
	"net/http"

	"github.com/dashhive/p2pwebproxy"
)

func main() {
	bindAddr := ":8080"
	log.Printf("Listening on %s", bindAddr)
	http.HandleFunc("/ws", p2pwebproxy.Handler)
	log.Fatal(http.ListenAndServe(bindAddr, nil))
}
