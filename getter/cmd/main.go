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

	"github.com/de4et/your-load/getter/internal/downloader"
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
		"https://stream.telko.ru/Pir5a_poz6_cam1/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam1/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam2/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam2/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam3/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam3/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam4/tracks-v1/index.fmp4.m3u8",
		"https://stream.telko.ru/Pir5a_poz6_cam4/tracks-v1/index.fmp4.m3u8",
	}

	downloaders := make([]downloader.StreamDownloader, len(urlsToDownload))
	ctx := context.Background()
	for i, v := range urlsToDownload {
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

	// const url = "https://live.hdontap.com/hls/hosb5/dollywood-eagles-aviary1-overlook_aef.stream/playlist.m3u8"
	// const url = "https://videos-3.earthcam.com/fecnetwork/3916.flv/chunklist_w294512582.m3u8?t=WHjF/VTcRlQ3Mh4Qco7ulOaxaZCUGiFHttf4YYxGU1TkioPqphoAzOc/ROODzHnP&td=202503142001"
	// const url1 = "https://stream.telko.ru/Pir5a_poz6_cam4/tracks-v1/index.fmp4.m3u8"
	// d := downloader.NewHLSStreamDownloader(url, rateInSeconds, 2)
	// d1 := downloader.NewHLSStreamDownloader(url1, rateInSeconds, 2)

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// d.Start(ctx)
	// d1.Start(ctx)

	// pts := int64(0)
	// log.Println("starting")
	// for {
	// 	img, err := d.Get()
	// 	if err != nil {
	// 		fmt.Printf("err: %v", err)
	// 		break
	// 	}

	// 	log.Printf("saving to file %d.jpg\n", pts)
	// 	err = saveToFile(img, pts)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	img, err = d1.Get()
	// 	if err != nil {
	// 		fmt.Printf("err: %v", err)
	// 		break
	// 	}

	// 	log.Printf("saving to file %d.jpg\n", pts+100000000000)
	// 	err = saveToFile(img, pts+100000000000)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	pts++
	// }
	wg.Wait()
	fmt.Println("Exiting...")
}
