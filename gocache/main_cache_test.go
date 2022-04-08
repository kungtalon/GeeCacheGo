package go_cache

import (
	"errors"
	"log"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
	"Zane": "578",
}

func TestGroup_Get(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("score", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[Callback] search key ", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, errors.New(key + " not exists")
		}))

	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to load value from db")
		}
		if _, err := gee.Get(k); err != nil {
			t.Fatalf("failed to get value from cache")
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but got %s instead", view)
	}
}
