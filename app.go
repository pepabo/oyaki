package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/h2non/bimg"
	"github.com/pepabo/oyaki/proxy"
	"github.com/pepabo/oyaki/router"
)

// App はアプリケーションのメイン構造体です
type App struct {
	config      proxy.Config
	proxyServer *proxy.ProxyServer
	router      *router.ExtensionRouter
}

// NewApp は新しいアプリケーションインスタンスを作成します
func NewApp() *App {
	return &App{}
}

// LoadConfig は環境変数から設定を読み込みます
func (a *App) LoadConfig() error {
	orgScheme := os.Getenv("OYAKI_ORIGIN_SCHEME")
	orgHost := os.Getenv("OYAKI_ORIGIN_HOST")
	if orgScheme == "" {
		orgScheme = "https"
	}

	quality := 90
	if q := os.Getenv("OYAKI_QUALITY"); q != "" {
		if parsed, err := strconv.Atoi(q); err == nil {
			quality = parsed
		}
	}

	timeout := 30 * time.Second
	if t := os.Getenv("OYAKI_TIMEOUT"); t != "" {
		if duration, err := time.ParseDuration(t); err == nil {
			timeout = duration
		}
	}

	a.config = proxy.Config{
		OriginURL: orgScheme + "://" + orgHost,
		Quality:   quality,
		Timeout:   timeout,
	}

	return nil
}

// InitializeVips はlibvipsを初期化します
func (a *App) InitializeVips() {
	bimg.Initialize()

	// キャッシュを無効化してメモリリークを防ぐ
	bimg.VipsCacheSetMax(0)
	bimg.VipsCacheSetMaxMem(0)
}

// ShutdownVips はlibvipsをシャットダウンします
func (a *App) ShutdownVips() {
	bimg.Shutdown()
}

// StartPprofServer はpprofサーバーをバックグラウンドで起動します
func (a *App) StartPprofServer() {
	go func() {
		log.Println("starting pprof server on localhost:6060")
		if err := http.ListenAndServe("127.0.0.1:6060", nil); err != nil {
			log.Printf("pprof server error: %v\n", err)
		}
	}()
}

// SetupRouter はルーターを設定します
func (a *App) SetupRouter() {
	a.proxyServer = proxy.NewProxyServer(a.config)

	a.router = router.NewExtensionRouter()
	a.router.HandleFunc(".webp", proxy.HandleWebP(a.proxyServer))
	a.router.HandleFunc(".jpg", proxy.HandleJPEG(a.proxyServer))
	a.router.HandleFunc(".jpeg", proxy.HandleJPEG(a.proxyServer))
	a.router.HandleFunc("", proxy.HandlePassthrough(a.proxyServer)) // デフォルト
}

// Run はHTTPサーバーを起動してアプリケーションを実行します
func (a *App) Run() error {
	log.Printf("starting oyaki %s\n", getVersion())
	log.Println("starting HTTP server on :8080")

	if err := http.ListenAndServe(":8080", a.router); err != nil {
		return err
	}

	return nil
}
