module github.com/allegro/bigcache-bench

go 1.13

require (
	github.com/allegro/bigcache/v2 v2.1.3
	github.com/coocood/freecache v1.1.0
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/polaris1119/cache v0.0.0-00010101000000-000000000000
)

replace github.com/polaris1119/cache => ../cache-example
