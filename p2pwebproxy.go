package p2pwebproxy

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sort"

	"github.com/coder/websocket"
)

var allowedList map[string][]string

var defaultAllowedPorts = []string{"80", "443"}

var validAccessToken = "your_secret_access_token"

func Init(rpcURL, user, pass string) (int, error) {
	var err error

	// TODO update occasionally
	allowedList, err = FetchAllowedList(rpcURL, user, pass)
	if err != nil {
		return 0, err
	}

	ips := getMapKeys(allowedList)
	subnetGroups := groupBySubnet24(ips)
	subnets := []string{}
	for subnet := range subnetGroups {
		subnets = append(subnets, subnet)
	}
	sortSubnets(subnets)
	for _, subnet := range subnets {
		ipList := subnetGroups[subnet]
		ip := ipList[0]
		ports := allowedList[ip]
		if len(ipList) == 1 {
			fmt.Printf("   1: %s:%s\n", ip, ports[0])
			continue
		}

		fmt.Printf(" %3d: %s:%s\n", len(ipList), ip, ports[0])
		ipList = ipList[1:]
		for _, ip := range ipList {
			ports := allowedList[ip]
			fmt.Printf("      %s:%s\n", ip, ports[0])
		}
	}

	return len(allowedList), nil
}

func groupBySubnet24(ips []string) map[string][]string {
	subnetMap := make(map[string][]string)

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			fmt.Printf("Invalid IP address: %s\n", ipStr)
			continue
		}
		// Convert IP to a /24 subnet by zeroing the last byte
		subnet := fmt.Sprintf("%d.%d.%d.0/24", ip[12], ip[13], ip[14])
		subnetMap[subnet] = append(subnetMap[subnet], ipStr)
	}

	return subnetMap
}

func getMapKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// func sortIPs(ips []string) []string {
//      sort.Slice(ips, func(i, j int) bool {
//              ip1 := net.ParseIP(ips[i])
//              ip2 := net.ParseIP(ips[j])
//              return bytesCompare(ip1.To16(), ip2.To16())
//      })
//      return ips
// }

func sortSubnets(subnets []string) []string {
	sort.Slice(subnets, func(i, j int) bool {
		_, net1, err1 := net.ParseCIDR(subnets[i])
		_, net2, err2 := net.ParseCIDR(subnets[j])

		if err1 != nil || err2 != nil {
			fmt.Printf("Invalid subnet: %s or %s\n", subnets[i], subnets[j])
			return false
		}

		return bytesCompare(net1.IP.To16(), net2.IP.To16())
	})
	return subnets
}

func bytesCompare(b1, b2 []byte) bool {
	for i := range b1 {
		if b1[i] < b2[i] {
			return true
		} else if b1[i] > b2[i] {
			return false
		}
	}
	return false
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
	addCORS(w, r)

	query := r.URL.Query()
	accessToken := query.Get("access_token")
	hostname := query.Get("hostname")
	port := query.Get("port")

	if false {
		user, err := authorize(accessToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		fmt.Println("dummy user", user)
	}
	fmt.Println("AUTHORIZATION IS TURNED OFF")

	if !isValidHostnamePort(hostname, port) {
		fmt.Printf("not in mn list: %s:%s\n", hostname, port)
		http.Error(w, "Forbidden: '%s:%s' not in mn list", http.StatusForbidden)
		return
	}

	wsconn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Required due to https://pkg.go.dev/github.com/coder/websocket#AcceptOptions
		// (it's about restricting Origin verification, which we actively don't want)
		InsecureSkipVerify: true,
	})
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

func AddCORSHandler(w http.ResponseWriter, r *http.Request) {
	addCORS(w, r)
	w.WriteHeader(http.StatusOK)
}

func addCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("origin")
	fmt.Printf("Origin 1: '%s'\n", origin)
	if len(origin) == 0 {
		host := r.Host
		if len(host) > 0 {
			origin = "https://" + host
		} else {
			origin = "http://localhost"
		}
	}
	fmt.Printf("Origin 2: '%s'\n", origin)

	w.Header().Set("Access-Control-Allow-Origin", origin) // Replace with your desired origin
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
		_ = closer.Close()
	}
	if closer, ok := dst.(io.Closer); ok {
		_ = closer.Close()
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
