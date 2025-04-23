//go:build cgo

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	_ "net/http/pprof"
	"time"

	"github.com/de4et/your-load/app/internal/getter"
	"github.com/de4et/your-load/app/internal/getter/checker"
	store "github.com/de4et/your-load/app/internal/getter/imagestore"
	"github.com/de4et/your-load/app/internal/getter/queue"
	"github.com/de4et/your-load/app/internal/repository/worker/maprep"
	getterservice "github.com/de4et/your-load/app/internal/service/getter"
	workerservice "github.com/de4et/your-load/app/internal/service/worker"
	"github.com/de4et/your-load/app/internal/worker/processor"
)

const (
	rateInSeconds           = 3
	handlerPeriodInSeconds  = 20
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

		"https://hd-auth.skylinewebcams.com/live.m3u8?a=36inma9414k4ih31j3112to445",
		"https://hd-auth.skylinewebcams.com/live.m3u8?a=36inma9414k4ih31j3112to445",

		"https://videos-3.earthcam.com/fecnetwork/8584.flv/chunklist_w1535994866.m3u8?t=mtYDpYsKJOUUjif3uNlVmRBtW72sY/cYyxmUBcEY0oCIp9kxki+FocK6/wLgkspt&td=202503251600",
		"https://videos-3.earthcam.com/fecnetwork/8584.flv/chunklist_w1535994866.m3u8?t=mtYDpYsKJOUUjif3uNlVmRBtW72sY/cYyxmUBcEY0oCIp9kxki+FocK6/wLgkspt&td=202503251600",

		"https://www.google.com/pal.m3u8",
		"https://www.google.com/pal.m3u8",

		"zxcvzxv",
	}

	s := store.NewFileImageStore()
	q := queue.NewSliceImageQueue()

	p := processor.NewStubProcessor()
	r := maprep.NewMapRepository()

	g := getterservice.NewGetterService(s, q)
	w := workerservice.NewWorkerService(s, q, p, r)

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
				CamID:         fmt.Sprintf("cam#%d", i),
				URL:           v,
				RateInSeconds: rateInSeconds,
				Type:          resp.ProtocolType,
			})
		}
	}

	ctxG, cancel := context.WithTimeout(context.Background(), time.Duration(40)*time.Second)
	defer cancel()
	for i, task := range tasks {
		log.Printf("Creating and adding job#%d", i)
		g.AddJob(time.Now().Add(time.Second*time.Duration(rand.Int()%15+handlerPeriodInSeconds)), task)
	}
	defer g.CloseAll()

	log.Printf("JOBS AMOUNT %d", g.Jobs())

	const numReceivers = 5
	ctxW, cancel := context.WithTimeout(context.Background(), time.Duration(recieverPeriodInSeconds)*time.Second)
	defer cancel()
	for i := 0; i < numReceivers; i++ {
		w.AddJob()
	}
	defer w.CloseAll()

	select {
	case <-ctxG.Done():
	case <-ctxW.Done():
	}

	log.Printf("JOBS AMOUNT %d", g.Jobs())
	log.Printf("Exiting...")
}
