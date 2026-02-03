package proxy

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"syscall"
	"time"
)

// HandlerFunc はプロキシハンドラーの関数型です
type HandlerFunc func(w http.ResponseWriter, r *http.Request, originURL string)

// ProxyServer はプロキシサーバーの設定と機能を提供します
type ProxyServer struct {
	client    *http.Client
	originURL string
	quality   int
}

// Config はProxyServerの設定を表します
type Config struct {
	OriginURL string
	Quality   int
	Timeout   time.Duration
}

// NewProxyServer は新しいProxyServerを作成します
func NewProxyServer(config Config) *ProxyServer {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &ProxyServer{
		client:    client,
		originURL: config.OriginURL,
		quality:   config.Quality,
	}
}

// GetOriginURL はOriginURLを返します
func (ps *ProxyServer) GetOriginURL() string {
	return ps.originURL
}

// GetQuality は画像品質設定を返します
func (ps *ProxyServer) GetQuality() int {
	return ps.quality
}

// FetchFromOrigin はオリジンサーバーからレスポンスを取得します
func (ps *ProxyServer) FetchFromOrigin(req *http.Request) (*http.Response, error) {
	return ps.client.Do(req)
}

// PrepareOriginRequest はオリジンサーバーへのリクエストを準備します
func (ps *ProxyServer) PrepareOriginRequest(r *http.Request, targetPath string) (*http.Request, error) {
	// オリジンURLを構築
	orgURL, err := url.Parse(ps.originURL + targetPath)
	if err != nil {
		log.Printf("Invalid origin URL: %v", err)
		return nil, err
	}

	// 新しいリクエストを作成
	req, err := http.NewRequest("GET", orgURL.String(), nil)
	if err != nil {
		log.Printf("Request creation failed: %v", err)
		return nil, err
	}

	// 必要なヘッダーをコピー
	req.Header.Set("User-Agent", "oyaki")

	if ifModifiedSince := r.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
		req.Header.Set("If-Modified-Since", ifModifiedSince)
	}

	if xff := r.Header.Get("X-Forwarded-For"); len(xff) > 1 {
		req.Header.Set("X-Forwarded-For", xff)
	}

	return req, nil
}

// SetResponseHeaders はレスポンスヘッダーを設定します
func (ps *ProxyServer) SetResponseHeaders(w http.ResponseWriter, orgRes *http.Response) {
	if lastModified := orgRes.Header.Get("Last-Modified"); lastModified != "" {
		w.Header().Set("Last-Modified", lastModified)
	} else {
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	}
}

// HandleError はエラーレスポンスを処理します
func (ps *ProxyServer) HandleError(w http.ResponseWriter, message string, statusCode int, err error) {
	http.Error(w, message, statusCode)
	log.Printf("%s: %v", message, err)
}

// CopyResponse はレスポンスボディをコピーします
func (ps *ProxyServer) CopyResponse(w http.ResponseWriter, src io.Reader) error {
	_, err := io.Copy(w, src)
	if err != nil {
		// クライアント側の接続切断は無視
		if !errors.Is(err, syscall.EPIPE) {
			log.Printf("Write response failed: %v", err)
			return err
		}
	}
	return nil
}
