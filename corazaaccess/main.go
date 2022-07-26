package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
)

const (
	httpStatusBlocked int = 403
	httpStatusError   int = 401
)

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

func main() {
	waf, err := coraza.NewWAF(coraza.NewWAFConfig().WithRequestBodyAccess(coraza.NewRequestBodyConfig().WithInMemoryLimit(100)).WithDirectivesFromFile("/src/coraza.conf").WithDirectivesFromFile("/src/coreruleset/crs-setup.conf.example").WithDirectivesFromFile("/src/coreruleset/rules/*.conf"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Starting POC")
	if err := http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("New transaction request")
		id := r.Header.Get("X-Coraza-ID")
		uri := r.Header.Get("X-Coraza-URL")
		if id == "" {
			http.Error(w, "X-Coraza-ID cannot be empty", httpStatusError)
			fmt.Println("X-Coraza-ID cannot be empty")
			return
		}
		if uri == "" {
			http.Error(w, "X-Coraza-URL cannot be empty", httpStatusError)
			fmt.Println("X-Coraza-URL cannot be empty")
		}
		u, err := url.Parse(uri)
		if err != nil {
			http.Error(w, err.Error(), httpStatusError)
			fmt.Println(err.Error())
		}
		*r.URL = *u

		for _, h := range []string{"X-Coraza-ID", "X-Coraza-URL"} {
			r.Header.Del(h)
		}
		tx := waf.NewTransactionWithID(id)
		defer func() {
			tx.ProcessLogging()
			tx.Close()
		}()
		w.Header().Set("X-Coraza-Id", id)
		if it, err := processRequest(tx, r); err != nil {
			http.Error(w, err.Error(), httpStatusError)
			fmt.Println("Request error:", err)
		} else if it != nil {
			w.WriteHeader(httpStatusBlocked)
			fmt.Fprint(w, it.RuleID)
			fmt.Println("Request blocked")
		} else if _, err := w.Write([]byte("ok")); err != nil {
			fmt.Println(err)
			fmt.Println("Response error")
			return
		}
		fmt.Printf("Transaction %s ok\n", id)
	})); err != nil {
		panic(err)
	}
}

func processRequest(tx types.Transaction, req *http.Request) (*types.Interruption, error) {
	var (
		client string
		cport  int
	)
	// IMPORTANT: Some http.Request.RemoteAddr implementations will not contain port or contain IPV6: [2001:db8::1]:8080
	idx := strings.LastIndexByte(req.RemoteAddr, ':')
	if idx != -1 {
		client = req.RemoteAddr[:idx]
		cport, _ = strconv.Atoi(req.RemoteAddr[idx+1:])
	}

	var in *types.Interruption
	// There is no socket access in the request object, so we neither know the server client nor port.
	tx.ProcessConnection(client, cport, "", 0)
	tx.ProcessURI(req.URL.String(), req.Method, req.Proto)
	for k, vr := range req.Header {
		for _, v := range vr {
			tx.AddRequestHeader(k, v)
		}
	}
	// Host will always be removed from req.Headers(), so we manually add it
	if req.Host != "" {
		tx.AddRequestHeader("Host", req.Host)
	}

	in = tx.ProcessRequestHeaders()
	if in != nil {
		return in, nil
	}
	if req.Body != nil && req.Body != http.NoBody {
		_, err := io.Copy(tx.RequestBodyWriter(), req.Body)
		if err != nil {
			return tx.Interruption(), err
		}
		_ = req.Body.Close()

		reader, err := tx.RequestBodyReader()
		if err != nil {
			return tx.Interruption(), err
		}
		// req.Body is transparently reinizialied with a new io.ReadCloser.
		// The http handler will be able to read it.
		// Prior to Go 1.19 NopCloser does not implement WriterTo if the reader implements it.
		// - https://github.com/golang/go/issues/51566
		// - https://tip.golang.org/doc/go1.19#minor_library_changes
		// This avoid errors like "failed to process request: malformed chunked encoding" when
		// using io.Copy
		// In Go 1.19 we just do `req.Body = io.NopCloser(reader)`
		if rwt, ok := reader.(io.WriterTo); ok {
			req.Body = struct {
				io.Reader
				io.WriterTo
				io.Closer
			}{reader, rwt, nopCloser{}}
		} else {
			req.Body = struct {
				io.Reader
				io.Closer
			}{reader, nopCloser{}}
		}
	}

	return tx.ProcessRequestBody()
}
