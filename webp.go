package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"

	"github.com/disintegration/imaging"
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

func convWebp(src io.Reader, params []string) (*bytes.Buffer, error) {
	f, err := os.CreateTemp("/tmp", "")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	defer os.Remove(f.Name())

	img, err := imaging.Decode(src, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	if err := imaging.Encode(f, img, imaging.JPEG); err != nil {
		return nil, err
	}

	params = append(params, "-quiet", "-mt", "-jpeg_like", f.Name(), "-o", "-")
	out, err := exec.Command("cwebp", params...).Output()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return bytes.NewBuffer(out), nil
}
