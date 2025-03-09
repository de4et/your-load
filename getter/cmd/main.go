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

	"github.com/aler9/gortsplib/pkg/h264"
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
	// setup client
	c := &gohlslib.Client{
		URI: "https://dai.google.com/linear/hls/pa/event/Sid4xiTQTkCT1SLu6rjUSQ/stream/b11233fc-f47d-4bdd-89f4-ff48267474d0:GRQ/variant/df9d5f9fc8201f0878fd4a77927eea3b/bandwidth/3016996.m3u8",
	}

	c.OnRequest = func(req *http.Request) {
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0")
	}
	// called when tracks are parsed
	c.OnTracks = func(tracks []*gohlslib.Track) error {
		// find the H264 track
		track := findH264Track(tracks)
		if track == nil {
			return fmt.Errorf("H264 track not found")
		}
		fmt.Println("ON TRACK")

		// create the H264 decoder
		frameDec := &h264Decoder{}
		err := frameDec.initialize()
		if err != nil {
			return err
		}

		// if SPS and PPS are present into the track, send them to the decoder
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
				// convert NALUs into RGBA frames
				naluType := nalu[0] & 0x1F
				nt := h264.NALUType(naluType)
				log.Println("nalu type -- ", nt)
				fuHeader := nalu[1]
				start := (fuHeader & 0x80) != 0
				end := (fuHeader & 0x40) != 0
				if start {
					fmt.Printf("Start of fragmented NAL unit (type %d)\n", naluType)
				}
				if end {
					fmt.Printf("End of fragmented NAL unit (type %d)\n", naluType)
				}
				img, err := frameDec.decode(nalu)
				if err != nil {
					panic(err)
				}

				if secs < last+float64(rateInSeconds) {
					continue
				}

				log.Printf("actual: %.2f and %.2f", time.Since(t).Seconds(), secs)
				// wait for a frame
				if img == nil {
					continue
				}

				// convert frame to JPEG and save to file
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

	// start reading
	err := c.Start()
	if err != nil {
		panic(err)
	}
	defer c.Close()
	// wait for a fatal error
	panic(<-c.Wait())
}

// func goid() int {
// 	var buf [64]byte
// 	n := runtime.Stack(buf[:], false)
// 	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
// 	id, err := strconv.Atoi(idField)
// 	if err != nil {
// 		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
// 	}
// 	return id
// }
