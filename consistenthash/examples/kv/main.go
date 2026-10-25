package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/mjaow/experimental-projects/consistenthash"
)

func main() {
	listen := flag.String("listen", ":8080", "listen address")
	shards := flag.String("shards", "", "comma-separated shard URLs; omit to run a storage shard")
	flag.Parse()

	handler := newStore()
	if *shards != "" {
		var err error
		handler, err = newRouter(strings.Split(*shards, ","))
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, handler))
}

func newRouter(shards []string) (http.Handler, error) {
	ring := consistenthash.New(100)
	proxies := map[string]*httputil.ReverseProxy{}
	for _, shard := range shards {
		shard = strings.TrimSpace(shard)
		target, err := url.Parse(shard)
		if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") {
			return nil, fmt.Errorf("invalid shard URL: %q", shard)
		}
		ring.Add(shard)
		proxies[shard] = httputil.NewSingleHostReverseProxy(target)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		shard := ring.Lookup(r.PathValue("key"))
		if shard == "" {
			http.Error(w, "no shards configured", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("X-Shard", shard)
		proxies[shard].ServeHTTP(w, r)
	})
	return mux, nil
}

func newStore() http.Handler {
	values := map[string][]byte{}
	var mu sync.Mutex
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		value, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		values[r.PathValue("key")] = value
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /kv/{key}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		value, ok := values[r.PathValue("key")]
		mu.Unlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(value)
	})
	return mux
}
