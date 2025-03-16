//go:build cgo

package main

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	_ "net/http/pprof"
	"os"
	"strconv"
	"sync"

	"github.com/de4et/your-load/getter/pkg/checker"
	"github.com/de4et/your-load/getter/pkg/downloader"
)

const rateInSeconds = 3

func saveToFile(img image.Image, pref int, pts int64) error {
	// create file
	fname := "imgs/" + strconv.FormatInt(int64(pref), 10) + "_" + strconv.FormatInt(pts, 10) + ".jpg"
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
	urlsToDownload := []string{
		"https://live.hdontap.com/hls/hosb5/dollywood-eagles-aviary1-overlook_aef.stream/playlist.m3u8",
		"https://live.hdontap.com/hls/hosb5/dollywood-eagles-aviary1-overlook_aef.stream/playlist.m3u8",

		"https://stream.telko.ru/Pir5a_poz6_cam1/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam1/tracks-v1/index.fmp4.m3u8",

		"https://stream.telko.ru/Pir5a_poz6_cam2/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam2/tracks-v1/index.fmp4.m3u8",

		"https://stream.telko.ru/Pir5a_poz6_cam3/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam3/tracks-v1/index.fmp4.m3u8",

		"https://stream.telko.ru/Pir5a_poz6_cam4/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam4/tracks-v1/index.fmp4.m3u8",

		"https://www.google.com/pal.m3u8",
		"https://www.google.com/pal.m3u8",
	}

	checkedURLs := make([]string, 0, len(urlsToDownload))
	c := checker.NewChecker()
	for i, v := range urlsToDownload {
		log.Printf("Checking #%d - %s", i, v)
		resp, err := c.CheckURL(v)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		log.Printf("%+v", resp)
		if resp.ProtocolType == checker.ProtocolHLS {
			checkedURLs = append(checkedURLs, v)
		}
	}

	downloaders := make([]downloader.StreamDownloader, len(checkedURLs))
	ctx := context.Background()
	for i, v := range checkedURLs {
		downloaders[i] = downloader.NewHLSStreamDownloader(v, rateInSeconds, 2)
		downloaders[i].Start(ctx)
	}

	wg := sync.WaitGroup{}

	for i, v := range downloaders {
		wg.Add(1)
		go func(down downloader.StreamDownloader, prefInt int) {
			pts := int64(0)
			for {
				img, err := down.Get()
				if err != nil {
					fmt.Printf("err: %v", err)
					break
				}

				log.Printf("saving to file %d_%d.jpg\n", prefInt, pts)
				err = saveToFile(img, prefInt, pts)
				if err != nil {
					panic(err)
				}
				pts++
			}
			wg.Done()

		}(v, i)
	}
	wg.Wait()
	fmt.Println("Exiting...")
}
