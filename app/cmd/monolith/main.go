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
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/de4et/your-load/app/internal/getter"
	"github.com/de4et/your-load/app/internal/getter/checker"
	"github.com/de4et/your-load/app/internal/getter/downloader"
	store "github.com/de4et/your-load/app/internal/getter/imagestore"
	"github.com/de4et/your-load/app/internal/pkg/queue"
)

const (
	rateInSeconds           = 3
	handlerPeriodInSeconds  = 10
	recieverPeriodInSeconds = 30
)

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

	s := store.NewMapImageStore()
	q := queue.NewSliceImageQueue()
	g := getter.NewGetter(s, q)

	tasks := make([]getter.Task, 0)
	for i, v := range urlsToDownload {
		log.Printf("Checking #%d - %s", i, v)
		resp, err := g.Check(v)
		if err != nil {
			log.Printf("ERROR: %v", err)
			continue
		}

		log.Printf("%+v", resp)
		if resp.ProtocolType == checker.ProtocolHLS {
			tasks = append(tasks, getter.Task{
				CamID: fmt.Sprintf("cam#%d", i),
				URL:   v,
			})
		}
	}

	for i, task := range tasks {
		log.Printf("Creating and adding job#%d", i)
		downloader := downloader.NewHLSStreamDownloader(task.URL, rateInSeconds, 2)
		job := getter.NewJob(time.Now().Add(time.Second*time.Duration(handlerPeriodInSeconds)), task, downloader)
		g.AddJob(job)
	}
	defer g.CloseAll()

	log.Printf("JOBS AMOUNT %d", g.Jobs())

	const numReceivers = 5
	rwg := sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(recieverPeriodInSeconds)*time.Second)
	defer cancel()
	for i := 0; i < numReceivers; i++ {
		rwg.Add(1)
		go recieveImage(ctx, "#"+strconv.FormatInt(int64(i), 10), &rwg, q, s)
	}

	rwg.Wait()

	log.Printf("JOBS AMOUNT %d", g.Jobs())
	log.Printf("Exiting...")
}

func recieveImage(ctx context.Context, name string, wg *sync.WaitGroup, q queue.ImageQueueGetter, s store.ImageStoreGetter) {
	const retryDelay = time.Duration(50) * time.Millisecond
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
