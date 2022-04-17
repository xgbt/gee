module main

go 1.17

require geeCache v0.0.0

require (
	geeCache/singleflight v0.0.0-00010101000000-000000000000 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

replace (
	geeCache => ./geeCache
	geeCache/singleflight => ./geeCache/singleflight
)
