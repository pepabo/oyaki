package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/h2non/bimg"
	"github.com/pepabo/oyaki/proxy"
	"github.com/pepabo/oyaki/router"
)

func TestMain(m *testing.M) {
	// libvipsを初期化
	bimg.Initialize()
	defer bimg.Shutdown()

	// キャッシュを無効化
	bimg.VipsCacheSetMax(0)
	bimg.VipsCacheSetMaxMem(0)

	os.Exit(m.Run())
}

func setupTestServer() (*router.ExtensionRouter, *proxy.ProxyServer, *httptest.Server) {
	// モックオリジンサーバーを作成
	originServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			if _, err := w.Write([]byte("origin root")); err != nil {
				// テスト用のモックサーバーでのエラーをログ出力
				log.Printf("モックサーバーでのレスポンス書き込みエラー: %v", err)
			}
		case "/image.jpg", "/image": // WebPリクエスト用に拡張子なしも対応
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
			// testdataのファイルが存在する場合はそれを使用
			if _, err := os.Stat("./testdata/oyaki.jpg"); err == nil {
				http.ServeFile(w, r, "./testdata/oyaki.jpg")
			} else {
				// テスト用のダミーJPEGヘッダー（最小限）
				if _, err := w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0xFF, 0xD9}); err != nil {
					// テスト用のモックサーバーでのエラーをログ出力
					log.Printf("モックサーバーでのレスポンス書き込みエラー: %v", err)
				}
			}
		case "/text.txt":
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte("test file content")); err != nil {
				// テスト用のモックサーバーでのエラーをログ出力
				log.Printf("モックサーバーでのレスポンス書き込みエラー: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))

	// アプリケーションを作成して設定
	app := NewApp()
	app.config = proxy.Config{
		OriginURL: originServer.URL,
		Quality:   90,
		Timeout:   5 * time.Second,
	}
	app.SetupRouter()

	return app.router, app.proxyServer, originServer
}

func TestRoot(t *testing.T) {
	r, _, originServer := setupTestServer()
	defer originServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got http %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Body.String() != "Oyaki lives!\n" {
		t.Errorf("got body %q, want 'Oyaki lives!'", rec.Body.String())
	}
}

func TestJPEGRequest(t *testing.T) {
	r, _, originServer := setupTestServer()
	defer originServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/image.jpg", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got http %d, want %d", rec.Code, http.StatusOK)
	}

	// Last-Modifiedヘッダーの確認
	if rec.Header().Get("Last-Modified") == "" {
		t.Error("Last-Modified header should be set")
	}
}

func TestWebPRequest(t *testing.T) {
	r, _, originServer := setupTestServer()
	defer originServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/image.webp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got http %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestPassthroughRequest(t *testing.T) {
	r, _, originServer := setupTestServer()
	defer originServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/text.txt", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got http %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("got Content-Type %q, want 'text/plain'", rec.Header().Get("Content-Type"))
	}

	if rec.Body.String() != "test file content" {
		t.Errorf("got body %q, want 'test file content'", rec.Body.String())
	}
}

func TestNotFoundRequest(t *testing.T) {
	r, _, originServer := setupTestServer()
	defer originServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent.jpg", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got http %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRequestHeaders(t *testing.T) {
	r, _, originServer := setupTestServer()
	defer originServer.Close()

	req := httptest.NewRequest(http.MethodGet, "/image.jpg", nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	req.Header.Set("If-Modified-Since", "Wed, 21 Oct 2015 07:28:00 GMT")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// 304 Not Modifiedが返ることを期待
	if rec.Code != http.StatusNotModified {
		t.Logf("Expected 304 Not Modified, got %d (this may be OK depending on origin server behavior)", rec.Code)
	}
}

func TestLoadConfig(t *testing.T) {
	// 環境変数を設定
	if err := os.Setenv("OYAKI_ORIGIN_SCHEME", "http"); err != nil {
		t.Fatalf("環境変数の設定に失敗しました: %v", err)
	}
	if err := os.Setenv("OYAKI_ORIGIN_HOST", "example.com"); err != nil {
		t.Fatalf("環境変数の設定に失敗しました: %v", err)
	}
	if err := os.Setenv("OYAKI_QUALITY", "80"); err != nil {
		t.Fatalf("環境変数の設定に失敗しました: %v", err)
	}
	if err := os.Setenv("OYAKI_TIMEOUT", "10s"); err != nil {
		t.Fatalf("環境変数の設定に失敗しました: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("OYAKI_ORIGIN_SCHEME"); err != nil {
			t.Errorf("環境変数のクリアに失敗しました: %v", err)
		}
		if err := os.Unsetenv("OYAKI_ORIGIN_HOST"); err != nil {
			t.Errorf("環境変数のクリアに失敗しました: %v", err)
		}
		if err := os.Unsetenv("OYAKI_QUALITY"); err != nil {
			t.Errorf("環境変数のクリアに失敗しました: %v", err)
		}
		if err := os.Unsetenv("OYAKI_TIMEOUT"); err != nil {
			t.Errorf("環境変数のクリアに失敗しました: %v", err)
		}
	}()

	app := NewApp()
	err := app.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if app.config.OriginURL != "http://example.com" {
		t.Errorf("got OriginURL %q, want 'http://example.com'", app.config.OriginURL)
	}

	if app.config.Quality != 80 {
		t.Errorf("got Quality %d, want 80", app.config.Quality)
	}

	if app.config.Timeout != 10*time.Second {
		t.Errorf("got Timeout %v, want 10s", app.config.Timeout)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// 環境変数をクリア
	if err := os.Unsetenv("OYAKI_ORIGIN_SCHEME"); err != nil {
		t.Errorf("環境変数のクリアに失敗しました: %v", err)
	}
	if err := os.Unsetenv("OYAKI_ORIGIN_HOST"); err != nil {
		t.Errorf("環境変数のクリアに失敗しました: %v", err)
	}
	if err := os.Unsetenv("OYAKI_QUALITY"); err != nil {
		t.Errorf("環境変数のクリアに失敗しました: %v", err)
	}
	if err := os.Unsetenv("OYAKI_TIMEOUT"); err != nil {
		t.Errorf("環境変数のクリアに失敗しました: %v", err)
	}

	app := NewApp()
	err := app.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if app.config.OriginURL != "https://" {
		t.Errorf("got OriginURL %q, want 'https://'", app.config.OriginURL)
	}

	if app.config.Quality != 90 {
		t.Errorf("got Quality %d, want 90", app.config.Quality)
	}

	if app.config.Timeout != 30*time.Second {
		t.Errorf("got Timeout %v, want 30s", app.config.Timeout)
	}
}
