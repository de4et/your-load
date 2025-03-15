package checker

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

type Checker struct {
}

type ProtocolType int

const (
	ProtocolHLS ProtocolType = iota
)

func (p ProtocolType) String() string {
	switch p {
	case ProtocolHLS:
		return "HLS"
	default:
		return "unknown"
	}
}

var ErrMethodNotAllowed = fmt.Errorf("Method not allowed")
var ErrBadStatusCode = fmt.Errorf("Status code is not 200")

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:134.0) Gecko/20100101 Firefox/134.0)"

func NewChecker() *Checker {
	return &Checker{}
}

type CheckerResponse struct {
	ProtocolType ProtocolType
}

func (c *Checker) CheckURL(uri string) (CheckerResponse, error) {
	b, err := c.checkHLS(uri)
	if err != nil {
		return CheckerResponse{}, err
	}
	if b {
		return CheckerResponse{
			ProtocolType: ProtocolHLS,
		}, nil
	}

	return CheckerResponse{}, ErrMethodNotAllowed
}

func (c *Checker) checkHLS(uri string) (bool, error) {
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("User-Agent", defaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, ErrBadStatusCode
	}

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "vnd.apple.mpegurl") || strings.Contains(contentType, "x-mpegurl") {
		return true, nil
	}

	filename := getFilenameFromURL(uri)
	if strings.Contains(filename, ".m3u8") {
		return true, nil
	}

	return false, nil
}

func getFilenameFromURL(uri string) string {
	fileNameWithExt := path.Base(uri)
	fileNameWithExt = strings.Split(fileNameWithExt, "?")[0]
	fileNameWithExt = strings.Split(fileNameWithExt, "#")[0]
	return uri
}
