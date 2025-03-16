package downloader

import (
	"context"
	"fmt"
	"image"
	"log"
	"net/http"
	"time"

	"github.com/bluenviron/gohlslib/v2"
	"github.com/bluenviron/gohlslib/v2/pkg/codecs"
)

var (
	DefaultHeaders = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	}

	ErrNoH264Track = fmt.Errorf("H264 track not found")
)

const defaultRetriesAfterError = 3
const defaultImageChanSize = 3

type HLSStremaDownloader struct {
	URI     string
	Rate    float64
	Headers map[string]string

	client    *gohlslib.Client
	ctx       context.Context
	imageChan chan image.Image
	errChan   chan error

	retriesAfterError int
	retries           int
}

func NewHLSStreamDownloader(uri string, rate float64, retriesAfterError int) *HLSStremaDownloader {
	if retriesAfterError < 0 {
		retriesAfterError = defaultRetriesAfterError
	}
	return &HLSStremaDownloader{
		URI:               uri,
		Rate:              rate,
		imageChan:         make(chan image.Image, defaultImageChanSize),
		errChan:           make(chan error),
		retriesAfterError: retriesAfterError,
	}
}

func (sd *HLSStremaDownloader) Get() (image.Image, error) {
	select {
	case res := <-sd.imageChan:
		return res, nil

	case err := <-sd.errChan:
		return nil, err
	}
}

func (sd *HLSStremaDownloader) Start(ctx context.Context) error {
	sd.ctx = ctx

	sd.client = &gohlslib.Client{
		URI: sd.URI,
	}

	sd.client.OnRequest = func(req *http.Request) {
		for k, v := range DefaultHeaders {
			req.Header.Add(k, v)
		}
	}

	sd.client.OnTracks = func(tracks []*gohlslib.Track) error {
		track := findH264Track(tracks)
		if track == nil {
			return ErrNoH264Track
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

		t := time.Now()
		last := 0.0
		sd.client.OnDataH26x(track, func(pts int64, dts int64, au [][]byte) {
			secs := float64(pts) / float64(track.ClockRate)

			for _, nalu := range au {
				img, err := frameDec.decode(nalu)
				if err != nil {
					sd.errChan <- err
					return
				}

				if secs < last+sd.Rate || img == nil {
					continue
				}

				log.Printf("actual: %.2f and %.2f", time.Since(t).Seconds(), secs)
				sd.imageChan <- img
				last = secs
			}
		})

		return nil
	}

	err := sd.client.Start()
	if err != nil {
		return err
	}

	// retry retriesAfterError times after any error encountered
	// mainly for "next segment not found" error
	go func() {
		err = <-sd.client.Wait()
		if err != nil {
			log.Printf("encountered error: %v", err)
			if sd.retries >= sd.retriesAfterError {
				return
			}

			sd.retries++
			sd.client.Close()
			err := sd.Start(ctx)
			if err != nil {
				sd.errChan <- err

			}
		}
	}()

	return nil
}

func findH264Track(tracks []*gohlslib.Track) *gohlslib.Track {
	for _, track := range tracks {
		if _, ok := track.Codec.(*codecs.H264); ok {
			return track
		}
	}
	return nil
}
