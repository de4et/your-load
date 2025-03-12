//go:build cgo

package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/bluenviron/gohlslib/v2"
	"github.com/bluenviron/gohlslib/v2/pkg/codecs"
)

const rateInSeconds = 3
const NALU_TYPE_NONIDR = 1

func findH264Track(tracks []*gohlslib.Track) *gohlslib.Track {
	for _, track := range tracks {
		if _, ok := track.Codec.(*codecs.H264); ok {
			return track
		}
	}
	return nil
}

func saveToFile(img image.Image, pts int64) error {
	// create file
	fname := "imgs/" + strconv.FormatInt(pts, 10) + ".jpg"
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.Println("saving", fname)

	// convert to jpeg
	return jpeg.Encode(f, img, &jpeg.Options{
		Quality: 100,
	})
}

func main() {
	c := &gohlslib.Client{
		URI: "https://minos.cloud.obrut.stream/movies/fb39b1dc117a53e7e03124c89d9b87517f55e4e8/b012ee55b49e5b3a65053574734ee2cb:2025031303/720.mp4:hls:manifest.m3u8",
	}

	c.OnRequest = func(req *http.Request) {
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0")
	}

	c.OnTracks = func(tracks []*gohlslib.Track) error {
		track := findH264Track(tracks)
		if track == nil {
			return fmt.Errorf("H264 track not found")
		}

		frameDec := &h264Decoder{}
		err := frameDec.initialize()
		if err != nil {
			return err
		}

		if track.Codec.(*codecs.H264).SPS != nil {
			frameDec.decode(track.Codec.(*codecs.H264).SPS)
		}
		if track.Codec.(*codecs.H264).PPS != nil {
			frameDec.decode(track.Codec.(*codecs.H264).PPS)
		}

		saveCount := 0

		t := time.Now()
		last := 0.0
		c.OnDataH26x(track, func(pts int64, dts int64, au [][]byte) {

			secs := float64(pts) / float64(track.ClockRate)

			for _, nalu := range au {
				img, err := frameDec.decode(nalu)
				if err != nil {
					panic(err)
				}

				if secs < last+float64(rateInSeconds) {
					continue
				}

				log.Printf("actual: %.2f and %.2f", time.Since(t).Seconds(), secs)
				if img == nil {
					continue
				}

				err = saveToFile(img, pts)
				if err != nil {
					panic(err)
				}

				last = secs
				saveCount++
				if saveCount == 100 {
					log.Println("Saved all, exiting...")
					os.Exit(1)
				}
			}
		})

		return nil
	}

	err := c.Start()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = <-c.Wait()
	if err != nil {
		log.Printf("ERROR - %v", err)
	}
	// panic(<-c.Wait())
}
