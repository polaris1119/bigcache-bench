package main

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/allegro/bigcache/v2"
	"github.com/coocood/freecache"
	"github.com/polaris1119/cache"
	"github.com/polaris1119/cache/lru"
)

var previousPause time.Duration

func gcPause() time.Duration {
	runtime.GC()
	var stats debug.GCStats
	debug.ReadGCStats(&stats)
	pause := stats.PauseTotal - previousPause
	previousPause = stats.PauseTotal
	return pause
}

const (
	entries   = 2000000
	valueSize = 100
	repeat    = 20
)

func main() {
	debug.SetGCPercent(5)
	fmt.Println("Number of entries: ", entries)
	fmt.Println("Number of repeats: ", repeat)

	fmt.Println("GC pause for startup: ", gcPause())

	stdMap()
	freeCache()
	tourCache()
	bigCache()

	fmt.Println("GC pause for warmup: ", gcPause())

	for i := 0; i < repeat; i++ {
		freeCache()
	}
	fmt.Println("GC pause for freecache: ", gcPause())
	for i := 0; i < repeat; i++ {
		bigCache()
	}
	fmt.Println("GC pause for bigcache: ", gcPause())
	for i := 0; i < repeat; i++ {
		tourCache()
	}
	fmt.Println("GC pause for tourcache: ", gcPause())
	for i := 0; i < repeat; i++ {
		stdMap()
	}
	fmt.Println("GC pause for map: ", gcPause())
}

func stdMap() {
	mapCache := make(map[string][]byte)
	for i := 0; i < entries; i++ {
		key, val := generateKeyValue(i, valueSize)
		mapCache[key] = val
	}
}

func freeCache() {
	freeCache := freecache.NewCache(entries * 200) //allocate entries * 200 bytes
	for i := 0; i < entries; i++ {
		key, val := generateKeyValue(i, valueSize)
		if err := freeCache.Set([]byte(key), val, 0); err != nil {
			fmt.Println("Error in set: ", err.Error())
		}
	}

	firstKey, _ := generateKeyValue(1, valueSize)
	checkFirstElement(freeCache.Get([]byte(firstKey)))

	if freeCache.OverwriteCount() != 0 {
		fmt.Println("Overwritten: ", freeCache.OverwriteCount())
	}
}

func tourCache() {
	tourCache := cache.NewTourCache(nil, lru.New(entries*valueSize, nil))

	for i := 0; i < entries; i++ {
		key, val := generateKeyValue(i, valueSize)
		tourCache.Set(key, val)
	}

	firstKey, expectedVal := generateKeyValue(1, valueSize)
	firstVal := tourCache.Get(firstKey)
	if !reflect.DeepEqual(firstVal, expectedVal) {
		fmt.Println("Wrong first element: ", firstVal)
	}
}

func bigCache() {
	config := bigcache.Config{
		Shards:             256,
		LifeWindow:         100 * time.Minute,
		MaxEntriesInWindow: entries,
		MaxEntrySize:       200,
		Verbose:            true,
	}

	bigcache, _ := bigcache.NewBigCache(config)
	for i := 0; i < entries; i++ {
		key, val := generateKeyValue(i, valueSize)
		bigcache.Set(key, val)
	}

	firstKey, _ := generateKeyValue(1, valueSize)
	checkFirstElement(bigcache.Get(firstKey))
}

func checkFirstElement(val []byte, err error) {
	_, expectedVal := generateKeyValue(1, valueSize)
	if err != nil {
		fmt.Println("Error in get: ", err.Error())
	} else if string(val) != string(expectedVal) {
		fmt.Println("Wrong first element: ", string(val))
	}
}

func generateKeyValue(index int, valSize int) (string, []byte) {
	key := fmt.Sprintf("key-%010d", index)
	fixedNumber := []byte(fmt.Sprintf("%010d", index))
	val := append(make([]byte, valSize-10), fixedNumber...)

	return key, val
}
