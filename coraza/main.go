package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/jptosso/coraza-waf/v2"
	"github.com/jptosso/coraza-waf/v2/seclang"
)

const (
	socketFile = "/var/run/coraza.sock"
)

var files = []string{
	"/coreruleset/coraza.conf",
	"/coreruleset/coreruleset/crs-setup.conf.example",
	"/coreruleset/coreruleset/rules/*.conf",
}

type ApiRequest struct {
}

func processRequest(c net.Conn) {
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())
	io.Copy(c, c)
	c.Close()
}

func main() {
	waf := coraza.NewWaf()
	waf.SetErrorLogCb(func(mr coraza.MatchedRule) {
		fmt.Printf("Error: %s\n", mr.ErrorLog(500))
	})
	p, _ := seclang.NewParser(waf)
	for _, f := range files {
		if err := p.FromFile(f); err != nil {
			panic(err)
		}
	}
	l, err := net.Listen("unix", socketFile)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer l.Close()

	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go processRequest(conn)
	}
}
