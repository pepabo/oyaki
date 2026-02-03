package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtensionRouter_HandleFunc(t *testing.T) {
	router := NewExtensionRouter()

	// テスト用ハンドラーを定義
	webpHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("webp handler")); err != nil {
			t.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}
	jpegHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("jpeg handler")); err != nil {
			t.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("default handler")); err != nil {
			t.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}

	// ハンドラーを登録
	router.HandleFunc(".webp", webpHandler)
	router.HandleFunc(".jpg", jpegHandler)
	router.HandleFunc(".jpeg", jpegHandler)
	router.HandleFunc("", defaultHandler)

	tests := []struct {
		name           string
		path           string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "WebP file",
			path:           "/image.webp",
			expectedBody:   "webp handler",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "JPEG file (.jpg)",
			path:           "/image.jpg",
			expectedBody:   "jpeg handler",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "JPEG file (.jpeg)",
			path:           "/image.jpeg",
			expectedBody:   "jpeg handler",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown extension",
			path:           "/file.txt",
			expectedBody:   "default handler",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No extension",
			path:           "/path",
			expectedBody:   "default handler",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Root path",
			path:           "/",
			expectedBody:   "default handler",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestExtensionRouter_NotFound(t *testing.T) {
	router := NewExtensionRouter()

	// デフォルトハンドラーを設定しない
	router.HandleFunc(".jpg", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("jpg handler")); err != nil {
			t.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	})

	req := httptest.NewRequest("GET", "/unknown.txt", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestExtensionRouter_EmptyPattern(t *testing.T) {
	router := NewExtensionRouter()

	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("default")); err != nil {
			t.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}

	// 空のパターンはデフォルトハンドラーとして扱われる
	router.HandleFunc("", defaultHandler)

	req := httptest.NewRequest("GET", "/any/path.unknown", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != "default" {
		t.Errorf("expected body %q, got %q", "default", body)
	}
}

// ベンチマークテスト
func BenchmarkExtensionRouter_ServeHTTP(b *testing.B) {
	router := NewExtensionRouter()

	webpHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("webp")); err != nil {
			b.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}
	jpegHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("jpeg")); err != nil {
			b.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}
	defaultHandler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("default")); err != nil {
			b.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}

	router.HandleFunc(".webp", webpHandler)
	router.HandleFunc(".jpg", jpegHandler)
	router.HandleFunc(".jpeg", jpegHandler)
	router.HandleFunc("", defaultHandler)

	req := httptest.NewRequest("GET", "/test.webp", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkExtensionRouter_MultipleExtensions(b *testing.B) {
	router := NewExtensionRouter()

	handler := func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("ok")); err != nil {
			b.Errorf("レスポンスの書き込みに失敗しました: %v", err)
		}
	}

	// 複数の拡張子を登録
	extensions := []string{".webp", ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".svg"}
	for _, ext := range extensions {
		router.HandleFunc(ext, handler)
	}
	router.HandleFunc("", handler)

	paths := []string{"/test.webp", "/test.jpg", "/test.png", "/test.unknown"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
