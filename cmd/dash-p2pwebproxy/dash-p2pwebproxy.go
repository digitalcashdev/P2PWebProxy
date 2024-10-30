package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dashhive/p2pwebproxy"
)

var (
	name    = "dash-p2pwebproxy"
	version = "0.0.0-dev"
	date    = "0001-01-01T00:00:00Z"
	commit  = "0000000"
)

func printVersion() {
	fmt.Printf("%s v%s %s (%s)\n", name, version, commit[:7], date)
	fmt.Printf("Copyright (C) 2024 AJ ONeal\n")
	fmt.Printf("Licensed under the MPL-2.0\n")
}

func printUsage() {
	fmt.Printf("%s v%s %s (%s)\n", name, version, commit[:7], date)
	fmt.Printf("\n")
	fmt.Printf("USAGE\n\tdash-p2pwebproxy --port <port>\n\n")
	fmt.Printf("EXAMPLES\n\tdash-p2pwebproxy --port 8080\n\tdash-p2pwebproxy --port 3000\n\n")
}

func main() {
	nArgs := len(os.Args)

	if nArgs >= 2 {
		arg := os.Args[1]
		subcmd := strings.TrimPrefix(arg, "-")
		if arg == "-V" || subcmd == "version" {
			printVersion()
			os.Exit(0)
			return
		}
		if subcmd == "help" {
			printUsage()
			os.Exit(0)
			return
		}
	}

	var wsPort int
	var testnet bool
	var rpcURL string

	defaultWSPort := 8080
	wsPortStr := os.Getenv("DASHD_P2P_WS_PORT")
	if len(wsPortStr) > 0 {
		defaultWSPort, _ = strconv.Atoi(wsPortStr)
		if defaultWSPort == 0 {
			defaultWSPort = 8080
		}
	}

	flag.IntVar(&wsPort, "port", defaultWSPort, "bind and listen for websockets on this port")
	flag.BoolVar(&testnet, "testnet", false, "only allow connections to testnet MNs (negated by --rpc-url)")
	flag.StringVar(&rpcURL, "rpc-url", "", "use a custom, authenticated RPC url, such as https://api:token@rpc.example.com/")
	flag.Parse()

	if len(rpcURL) == 0 {
		if testnet {
			rpcURL = "https://api:null@trpc.digitalcash.dev/"
		} else {
			rpcURL = "https://api:null@rpc.digitalcash.dev/"
		}
	}

	log.Printf("RPC URL: %s", rpcURL)
	if n, err := p2pwebproxy.Init(rpcURL, "", ""); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("    found %d MNs", n)
	}

	http.HandleFunc("OPTIONS /ws", p2pwebproxy.AddCORSHandler)
	http.HandleFunc("GET /ws", p2pwebproxy.Handler)
	addr := fmt.Sprintf(":%d", wsPort)

	log.Printf("Listening on %d", wsPort)
	log.Fatal(http.ListenAndServe(addr, nil))
}
