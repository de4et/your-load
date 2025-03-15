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

	"github.com/de4et/your-load/getter/internal/downloader"
)

const rateInSeconds = 3

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
	const url = "https://live.hdontap.com/hls/hosb5/dollywood-eagles-aviary1-overlook_aef.stream/playlist.m3u8"
	// const url = "https://videos-3.earthcam.com/fecnetwork/3916.flv/chunklist_w294512582.m3u8?t=WHjF/VTcRlQ3Mh4Qco7ulOaxaZCUGiFHttf4YYxGU1TkioPqphoAzOc/ROODzHnP&td=202503142001"
	// const url = "https://stream.telko.ru/Pir5a_poz6_cam4/tracks-v1/index.fmp4.m3u8"
	d := downloader.NewHLSStreamDownloader(url, rateInSeconds, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d.Start(ctx)

	pts := int64(0)
	log.Println("starting")
	for {
		img, err := d.Get()
		if err != nil {
			fmt.Printf("err: %v", err)
			break
		}

		log.Printf("saving to file %d.jpg\n", pts)
		err = saveToFile(img, pts)
		if err != nil {
			panic(err)
		}
		pts++
	}
	fmt.Println("Exiting...")
}
