package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/h2non/bimg"
)

func doWebp(req *http.Request) (*http.Response, error) {
	var orgRes *http.Response
	orgURL := req.URL
	newPath := orgURL.Path[:len(orgURL.Path)-len(".webp")]
	newOrgURL, err := url.Parse(fmt.Sprintf("%s://%s%s?%s", orgURL.Scheme, orgURL.Host, newPath, orgURL.RawQuery))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	newReq, err := http.NewRequest("GET", newOrgURL.String(), nil)
	newReq.Header = req.Header
	if err != nil {
		log.Println(err)
		return nil, err
	}
	orgRes, err = client.Do(newReq)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if orgRes.StatusCode != 200 && orgRes.StatusCode != 304 {
		log.Println(orgRes.Status)
		return nil, fmt.Errorf("origin response is not 200 or 304")
	}

	if orgRes == nil {
		return nil, fmt.Errorf("origin response is not found")
	}
	return orgRes, nil
}

func convWebp(src []byte, quality int) (*bytes.Buffer, error) {
	opts := bimg.Options{
		Type:         bimg.WEBP,
		Quality:      quality,
		NoAutoRotate: false,
		// NoAutoRotateはデフォルトでfalseで、勝手にrotateしてくれる

		// Safariなどでは、bimgによってEXIFの回転処理を実施したあとにブラウザ側でEXIFを読んで再度回転してしまうことがあるので、
		// EXIFは削除する
		StripMetadata: true,
	}
	webpImg, err := bimg.NewImage(src).Process(opts)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(webpImg), nil
}
