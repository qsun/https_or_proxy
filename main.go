package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/net/proxy"

	"github.com/qsun/go-socks5"
)

var (
	listenSpec = flag.String("listen", "127.0.0.1:3333", "listen host and port")
	socks5Spec = flag.String("socks5", "127.0.0.1:3334", "forward socks5 server")
)

type HTTPSOrProxyConnectHandler struct {
	forward proxy.Dialer
}

func (h HTTPSOrProxyConnectHandler) Connect(s *socks5.Server, c socks5.Conn, bufConn io.Reader, dest *socks5.AddrSpec, realDest *socks5.AddrSpec) (*net.TCPConn, error) {
	if dest.Port == 443 {
		log.Println("HTTPS connection")

		return s.Connect(s, c, bufConn, dest, realDest)
	} else {
		log.Println("forward connection")
		target := dest.IP.String() + ":" + fmt.Sprintf("%d", dest.Port)
		log.Println("Target: ", target)
		c, err := h.forward.Dial("tcp", target)
		if err != nil {
			return nil, err
		}

		return c.(*net.TCPConn), nil
	}
}

func main() {
	flag.Parse()
	log.Println("Started on: ", *listenSpec, "with forward: ", *socks5Spec)

	dialer, err := proxy.SOCKS5("tcp", *socks5Spec, nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Socks5 dialer initialization failed: %v", err)
	}

	h := HTTPSOrProxyConnectHandler{
		forward: dialer,
	}

	conf := &socks5.Config{
		ConnectHandler: h,
	}
	serv, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}

	if err = serv.ListenAndServe("tcp", *listenSpec); err != nil {
		log.Fatal(err)
	}
}
