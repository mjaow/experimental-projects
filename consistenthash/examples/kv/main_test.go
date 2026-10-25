package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShardedService(t *testing.T) {
	shards := make([]string, 3)
	for i := range shards {
		server := httptest.NewServer(newStore())
		t.Cleanup(server.Close)
		shards[i] = server.URL
	}
	router, err := newRouter(shards)
	if err != nil {
		t.Fatal(err)
	}
	owners := map[string]bool{}
	for i := 0; i < 60; i++ {
		path := fmt.Sprintf("/kv/user-%d", i)
		value := fmt.Sprintf("profile-%d", i)
		put := httptest.NewRecorder()
		router.ServeHTTP(put, httptest.NewRequest(http.MethodPut, path, strings.NewReader(value)))
		if put.Code != http.StatusNoContent {
			t.Fatalf("PUT %s: %d %s", path, put.Code, put.Body.String())
		}
		owner := put.Header().Get("X-Shard")
		owners[owner] = true
		get := httptest.NewRecorder()
		router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, path, nil))
		if get.Code != http.StatusOK || get.Body.String() != value || get.Header().Get("X-Shard") != owner {
			t.Fatalf("GET %s did not read its value from the same shard", path)
		}
		for _, shard := range shards {
			resp, err := http.Get(shard + path)
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if shard == owner {
				if resp.StatusCode != http.StatusOK || string(body) != value {
					t.Fatalf("owner %s did not store %s", shard, path)
				}
			} else if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("non-owner %s stored %s", shard, path)
			}
		}
	}
	if len(owners) != 3 {
		t.Fatalf("keys reached %d shards, want 3", len(owners))
	}
}
