//go:build cgo

package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/de4et/your-load/getter/internal/checker"
	"github.com/de4et/your-load/getter/internal/downloader"
	store "github.com/de4et/your-load/getter/internal/image"
	"github.com/de4et/your-load/getter/internal/queue"
)

const rateInSeconds = 3

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

		"zxcvzxv",
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
		defer downloaders[i].Close()
	}

	q := queue.NewSliceImageQueue()
	s := store.NewMapStore()

	wg := sync.WaitGroup{}
	for i, v := range downloaders {
		wg.Add(1)
		go func(down downloader.StreamDownloader, prefInt int) {
			defer wg.Done()
			pts := int64(0)
			camID := fmt.Sprintf("cam#%d", prefInt)
			ctx := context.TODO()
			for {
				resp, err := down.Get()
				if err != nil {
					if errors.Is(err, downloader.ErrClosed) {
						return
					}

					fmt.Printf("err: %v", err)
					break
				}

				name := fmt.Sprintf("%s_%d", camID, resp.Timestamp.UnixNano())
				uri, err := s.Add(ctx, resp.Image, name)
				if err != nil {
					panic(err)
				}
				log.Printf("saved to %s", uri)

				q.Add(ctx, queue.ImageQueueElement{
					Timestamp: resp.Timestamp,
					ImageURI:  uri,
					CamID:     camID,
				})
				pts++
			}
		}(v, i)
	}

	const numReceivers = 3
	rwg := sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()
	for i := 0; i < numReceivers; i++ {
		rwg.Add(1)
		go recieveImage(ctx, "#"+strconv.FormatInt(int64(i), 10), &rwg, q, s)
	}

	rwg.Wait()
	log.Printf("Exiting...")
}

func recieveImage(ctx context.Context, name string, wg *sync.WaitGroup, q queue.ImageQueueGetter, s store.ImageStoreGetter) {
	const retryDelay = time.Duration(100) * time.Millisecond
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Printf("CTX RECIEVER is closed")
			return
		default:
		}

		el, err := q.Get(ctx)
		if err != nil {
			if err == queue.ErrQueueIsEmpty {
				time.Sleep(retryDelay)
				continue
			}
		}

		img, err := s.Get(ctx, el.ImageURI)
		if err != nil {
			panic(err)
		}

		absName, err := saveToFile(img, el.ImageURI)
		if err != nil {
			panic(err)
		}

		log.Printf("%s ID-%s TS-%v AbsName-%s", name, el.CamID, el.Timestamp, absName)
	}
}

func saveToFile(img image.Image, name string) (string, error) {
	fname := "imgs/" + name + ".jpg"
	f, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.Println("saving", fname)

	absPath, err := filepath.Abs(fname)
	if err != nil {
		return "", err
	}

	return absPath, jpeg.Encode(f, img, &jpeg.Options{
		Quality: 100,
	})
}
