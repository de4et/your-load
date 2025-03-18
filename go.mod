module github.com/de4et/your-load

go 1.21.0

require github.com/bluenviron/gohlslib/v2 v2.1.3

require (
	github.com/abema/go-mp4 v1.4.1 // indirect
	github.com/asticode/go-astikit v0.30.0 // indirect
	github.com/asticode/go-astits v1.13.0 // indirect
	github.com/bluenviron/mediacommon v1.13.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
)

replace github.com/bluenviron/mediacommon/v2 => ./getter/vend/mediacommon/v2@v2.0.0

replace github.com/bluenviron/mediacommon => ./getter/vend/mediacommon@v1.14.0
