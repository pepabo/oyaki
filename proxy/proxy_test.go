package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewProxyServer(t *testing.T) {
	config := Config{
		OriginURL: "https://example.com",
		Quality:   80,
		Timeout:   10 * time.Second,
	}

	ps := NewProxyServer(config)

	if ps.GetOriginURL() != config.OriginURL {
		t.Errorf("expected OriginURL %s, got %s", config.OriginURL, ps.GetOriginURL())
	}

	if ps.GetQuality() != config.Quality {
		t.Errorf("expected Quality %d, got %d", config.Quality, ps.GetQuality())
	}
}

func TestNewProxyServer_DefaultTimeout(t *testing.T) {
	config := Config{
		OriginURL: "https://example.com",
		Quality:   90,
		// Timeout is not set
	}

	ps := NewProxyServer(config)

	// デフォルトタイムアウトが設定されることを確認（詳細な検証は難しいため、パニックしないことを確認）
	if ps == nil {
		t.Error("ProxyServer should not be nil")
	}
}

func TestProxyServer_PrepareOriginRequest(t *testing.T) {
	config := Config{
		OriginURL: "https://origin.example.com",
		Quality:   90,
	}
	ps := NewProxyServer(config)

	tests := []struct {
		name        string
		targetPath  string
		headers     map[string]string
		expectError bool
	}{
		{
			name:       "Basic path",
			targetPath: "/image.jpg",
			headers:    map[string]string{},
		},
		{
			name:       "With If-Modified-Since",
			targetPath: "/image.jpg",
			headers: map[string]string{
				"If-Modified-Since": "Wed, 21 Oct 2015 07:28:00 GMT",
			},
		},
		{
			name:       "With X-Forwarded-For",
			targetPath: "/image.jpg",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
		},
		{
			name:       "Complex path",
			targetPath: "/dir/subdir/image.webp?quality=80&format=webp",
			headers:    map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// オリジナルリクエストを作成
			originalReq := httptest.NewRequest("GET", "http://proxy.example.com"+tt.targetPath, nil)
			for key, value := range tt.headers {
				originalReq.Header.Set(key, value)
			}

			req, err := ps.PrepareOriginRequest(originalReq, tt.targetPath)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// URLの検証
			expectedURL := config.OriginURL + tt.targetPath
			if req.URL.String() != expectedURL {
				t.Errorf("expected URL %s, got %s", expectedURL, req.URL.String())
			}

			// User-Agentヘッダーの検証
			if req.Header.Get("User-Agent") != "oyaki" {
				t.Errorf("expected User-Agent 'oyaki', got %s", req.Header.Get("User-Agent"))
			}

			// 他のヘッダーの検証
			for key, expectedValue := range tt.headers {
				if key == "X-Forwarded-For" && len(expectedValue) <= 1 {
					// X-Forwarded-Forは長さが1以下の場合は設定されない
					continue
				}
				if req.Header.Get(key) != expectedValue {
					t.Errorf("expected header %s: %s, got %s", key, expectedValue, req.Header.Get(key))
				}
			}
		})
	}
}

func TestProxyServer_SetResponseHeaders(t *testing.T) {
	ps := NewProxyServer(Config{})

	tests := []struct {
		name                 string
		lastModified         string
		expectedLastModified bool
	}{
		{
			name:                 "With Last-Modified header",
			lastModified:         "Wed, 21 Oct 2015 07:28:00 GMT",
			expectedLastModified: true,
		},
		{
			name:                 "Without Last-Modified header",
			lastModified:         "",
			expectedLastModified: true, // デフォルト値が設定される
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックレスポンスを作成
			orgRes := &http.Response{
				Header: make(http.Header),
			}
			if tt.lastModified != "" {
				orgRes.Header.Set("Last-Modified", tt.lastModified)
			}

			w := httptest.NewRecorder()
			ps.SetResponseHeaders(w, orgRes)

			if tt.expectedLastModified {
				lastModified := w.Header().Get("Last-Modified")
				if lastModified == "" {
					t.Error("Last-Modified header should be set")
				}
				if tt.lastModified != "" && lastModified != tt.lastModified {
					t.Errorf("expected Last-Modified %s, got %s", tt.lastModified, lastModified)
				}
			}
		})
	}
}

func TestProxyServer_CopyResponse(t *testing.T) {
	ps := NewProxyServer(Config{})

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:  "Normal copy",
			input: "test content",
		},
		{
			name:  "Empty content",
			input: "",
		},
		{
			name:  "Large content",
			input: strings.Repeat("a", 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := strings.NewReader(tt.input)
			w := httptest.NewRecorder()

			err := ps.CopyResponse(w, src)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if w.Body.String() != tt.input {
				t.Errorf("expected body %q, got %q", tt.input, w.Body.String())
			}
		})
	}
}

func TestProxyServer_HandleError(t *testing.T) {
	ps := NewProxyServer(Config{})

	w := httptest.NewRecorder()
	testError := fmt.Errorf("test error")

	ps.HandleError(w, "Test message", http.StatusInternalServerError, testError)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	expectedBody := "Test message\n"
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, w.Body.String())
	}
}

// モックサーバーを使った統合テスト
func TestProxyServer_Integration(t *testing.T) {
	// モックオリジンサーバーを作成
	originServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/image.jpg" {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Header().Set("Last-Modified", "Wed, 21 Oct 2015 07:28:00 GMT")
			if _, err := w.Write([]byte("fake jpeg data")); err != nil {
				// テスト用のモックサーバーでのエラーをログ出力
				log.Printf("モックサーバーでのレスポンス書き込みエラー: %v", err)
			}
		} else {
			http.NotFound(w, r)
		}
	}))
	defer originServer.Close()

	config := Config{
		OriginURL: originServer.URL,
		Quality:   90,
		Timeout:   5 * time.Second,
	}
	ps := NewProxyServer(config)

	t.Run("Successful request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://proxy.example.com/image.jpg", nil)

		originReq, err := ps.PrepareOriginRequest(req, "/image.jpg")
		if err != nil {
			t.Fatalf("PrepareOriginRequest failed: %v", err)
		}

		resp, err := ps.FetchFromOrigin(originReq)
		if err != nil {
			t.Fatalf("FetchFromOrigin failed: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Errorf("レスポンスボディのクローズに失敗しました: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("ReadAll failed: %v", err)
		}

		if string(body) != "fake jpeg data" {
			t.Errorf("expected body 'fake jpeg data', got %q", string(body))
		}
	})
}
