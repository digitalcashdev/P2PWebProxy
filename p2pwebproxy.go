package p2pwebproxy

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/coder/websocket"
)

var allowedList map[string][]string

var defaultAllowedPorts = []string{"80", "443"}

var validAccessToken = "your_secret_access_token"

func init() {
	var err error

	// TODO update occasionally
	rpcURL := "https://rpc.digitalcash.dev/"
	allowedList, err = fetchAllowedList(rpcURL, "user", "pass")
	if err != nil {
		allowedList = map[string][]string{
			"localhost":   {"8080", "9090"},
			"example.com": {"80", "443"},
		}
	}
}

func authorize(accessToken string) (string, error) {
	isSame := secureCompare(accessToken, validAccessToken)
	if isSame {
		return "", errors.New("invalid token")
	}
	return "TODO", nil
}

func secureCompare(a, b string) bool {
	diff := subtle.ConstantTimeCompare([]byte(a), []byte(b))
	return diff == 1
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := r.URL.Query()
	accessToken := query.Get("access_token")
	hostname := query.Get("hostname")
	port := query.Get("port")

	user, err := authorize(accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Println("dummy user", user)

	if !isValidHostnamePort(hostname, port) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	wsconn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to websocket:", err)
		return
	}
	defer wsconn.Close(websocket.StatusInternalError, "Internal server error")

	host := fmt.Sprintf("%s:%s", hostname, port)
	tcpConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println("Failed to connect to TCP socket:", err)
		_ = wsconn.Write(ctx, websocket.MessageText, []byte("Failed to connect to TCP socket"))
		return
	}
	defer tcpConn.Close()

	// TODO rateLimit
	// https://github.com/coder/websocket/blob/e4379472fe1dfe70032ecc68fec08b1b3a8fc996/internal/examples/echo/server.go#L38

	byteCounter := &ByteCounter{}
	go pipeWithCounter(tcpConn, websocket.NetConn(ctx, wsconn, websocket.MessageBinary), byteCounter, "TCP to WebSocket")
	pipeWithCounter(websocket.NetConn(ctx, wsconn, websocket.MessageBinary), tcpConn, byteCounter, "WebSocket to TCP")

	wsconn.Close(websocket.StatusNormalClosure, "")
	log.Printf("Total bytes transferred: %d\n", byteCounter.Count)
}

func isValidHostnamePort(hostname, port string) bool {
	allowedPorts, exists := allowedList[hostname]
	if !exists {
		allowedPorts = defaultAllowedPorts
	}
	for _, p := range allowedPorts {
		if p == port {
			return true
		}
	}
	return false
}

func pipeWithCounter(src io.Reader, dst io.Writer, counter *ByteCounter, direction string) {
	_, err := io.Copy(io.MultiWriter(dst, counter), src)
	if err == io.EOF {
		log.Printf("Connection closed by remote side during %s\n", direction)
	} else {
		log.Printf("Error during %s: %v\n", direction, err)
	}

	log.Printf("Error during %s: %v\n", direction, err)
	if closer, ok := src.(io.Closer); ok {
		closer.Close()
	}
	if closer, ok := dst.(io.Closer); ok {
		closer.Close()
	}
}

type ByteCounter struct {
	Count int64
}

func (bc *ByteCounter) Write(p []byte) (int, error) {
	n := len(p)
	bc.Count += int64(n)
	return n, nil
}
