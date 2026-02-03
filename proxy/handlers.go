package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pepabo/oyaki/convert"
)

// HandleWebP はWebPファイルのリクエストを処理します
func HandleWebP(ps *ProxyServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.RequestURI()
		if path == "/" {
			if _, err := fmt.Fprintln(w, "Oyaki lives!"); err != nil {
				// レスポンス書き込みエラーは通常処理を続行不可
				return
			}
			return
		}

		// .webp拡張子を取り除いたパスでオリジンサーバーにリクエスト
		originalPath := strings.TrimSuffix(path, ".webp")
		req, err := ps.PrepareOriginRequest(r, originalPath)
		if err != nil {
			ps.HandleError(w, "Invalid origin URL", http.StatusBadRequest, err)
			return
		}

		// WebP用の特別なリクエスト処理
		orgRes, err := fetchWebPOrigin(ps, req)
		if err != nil {
			ps.HandleError(w, "Get origin failed", http.StatusForbidden, err)
			return
		}
		defer func() {
			if err := orgRes.Body.Close(); err != nil {
				log.Printf("レスポンスボディのクローズに失敗しました: %v", err)
			}
		}()

		// エラーレスポンスの処理
		if orgRes.StatusCode == http.StatusNotFound || orgRes.StatusCode == http.StatusForbidden {
			ps.HandleError(w, "Get origin failed", orgRes.StatusCode, fmt.Errorf("origin returned: %s", orgRes.Status))
			return
		}

		// レスポンスヘッダーを設定
		ps.SetResponseHeaders(w, orgRes)

		// 304 Not Modifiedの場合
		if orgRes.StatusCode == http.StatusNotModified {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// 200以外のレスポンス
		if orgRes.StatusCode != http.StatusOK {
			ps.HandleError(w, "Get origin failed", http.StatusBadGateway, fmt.Errorf("origin returned: %s", orgRes.Status))
			return
		}

		// Content-Typeチェック
		ct := orgRes.Header.Get("Content-Type")
		if ct != "image/jpeg" {
			// JPEG以外はそのまま返す
			w.Header().Set("Content-Type", ct)
			if cl := orgRes.Header.Get("Content-Length"); cl != "" {
				w.Header().Set("Content-Length", cl)
			}
			if err := ps.CopyResponse(w, orgRes.Body); err != nil {
				log.Printf("レスポンスのコピーに失敗しました: %v", err)
			}
			return
		}

		// JPEG→WebP変換処理
		handleJPEGToWebPConversion(w, orgRes, ps)
	}
}

// HandleJPEG はJPEGファイルのリクエストを処理します
func HandleJPEG(ps *ProxyServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.RequestURI()
		if path == "/" {
			if _, err := fmt.Fprintln(w, "Oyaki lives!"); err != nil {
				// レスポンス書き込みエラーは通常処理を続行不可
				return
			}
			return
		}

		req, err := ps.PrepareOriginRequest(r, path)
		if err != nil {
			ps.HandleError(w, "Invalid origin URL", http.StatusBadRequest, err)
			return
		}

		orgRes, err := ps.FetchFromOrigin(req)
		if err != nil {
			ps.HandleError(w, "Get origin failed", http.StatusForbidden, err)
			return
		}
		defer func() {
			if err := orgRes.Body.Close(); err != nil {
				log.Printf("レスポンスボディのクローズに失敗しました: %v", err)
			}
		}()

		// エラーレスポンスの処理
		if orgRes.StatusCode == http.StatusNotFound || orgRes.StatusCode == http.StatusForbidden {
			ps.HandleError(w, "Get origin failed", orgRes.StatusCode, fmt.Errorf("origin returned: %s", orgRes.Status))
			return
		}

		// レスポンスヘッダーを設定
		ps.SetResponseHeaders(w, orgRes)

		// 304 Not Modifiedの場合
		if orgRes.StatusCode == http.StatusNotModified {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// 200以外のレスポンス
		if orgRes.StatusCode != http.StatusOK {
			ps.HandleError(w, "Get origin failed", http.StatusBadGateway, fmt.Errorf("origin returned: %s", orgRes.Status))
			return
		}

		// Content-Typeチェック
		ct := orgRes.Header.Get("Content-Type")
		if ct != "image/jpeg" {
			// JPEG以外はそのまま返す
			w.Header().Set("Content-Type", ct)
			if cl := orgRes.Header.Get("Content-Length"); cl != "" {
				w.Header().Set("Content-Length", cl)
			}
			if err := ps.CopyResponse(w, orgRes.Body); err != nil {
				log.Printf("レスポンスのコピーに失敗しました: %v", err)
			}
			return
		}

		// JPEG最適化処理
		result, err := convert.ToJPEG(orgRes.Body, ps.GetQuality())
		if err != nil {
			ps.HandleError(w, "Image convert failed", http.StatusInternalServerError, err)
			return
		}
		defer result.Data.Reset()

		w.Header().Set("Content-Type", result.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(result.Data.Len()))

		if err := ps.CopyResponse(w, result.Data); err != nil {
			log.Printf("レスポンスのコピーに失敗しました: %v", err)
		}
	}
}

// HandlePassthrough はその他のファイルのリクエストを処理します
func HandlePassthrough(ps *ProxyServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.RequestURI()
		if path == "/" {
			if _, err := fmt.Fprintln(w, "Oyaki lives!"); err != nil {
				// レスポンス書き込みエラーは通常処理を続行不可
				return
			}
			return
		}

		req, err := ps.PrepareOriginRequest(r, path)
		if err != nil {
			ps.HandleError(w, "Invalid origin URL", http.StatusBadRequest, err)
			return
		}

		orgRes, err := ps.FetchFromOrigin(req)
		if err != nil {
			ps.HandleError(w, "Get origin failed", http.StatusForbidden, err)
			return
		}
		defer func() {
			if err := orgRes.Body.Close(); err != nil {
				log.Printf("レスポンスボディのクローズに失敗しました: %v", err)
			}
		}()

		// エラーレスポンスの処理
		if orgRes.StatusCode == http.StatusNotFound || orgRes.StatusCode == http.StatusForbidden {
			ps.HandleError(w, "Get origin failed", orgRes.StatusCode, fmt.Errorf("origin returned: %s", orgRes.Status))
			return
		}

		// レスポンスヘッダーを設定
		ps.SetResponseHeaders(w, orgRes)

		// 304 Not Modifiedの場合
		if orgRes.StatusCode == http.StatusNotModified {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// 200以外のレスポンス
		if orgRes.StatusCode != http.StatusOK {
			ps.HandleError(w, "Get origin failed", http.StatusBadGateway, fmt.Errorf("origin returned: %s", orgRes.Status))
			return
		}

		// ヘッダーをコピーしてそのまま転送
		ct := orgRes.Header.Get("Content-Type")
		cl := orgRes.Header.Get("Content-Length")

		w.Header().Set("Content-Type", ct)
		if cl != "" {
			w.Header().Set("Content-Length", cl)
		}

		if err := ps.CopyResponse(w, orgRes.Body); err != nil {
			log.Printf("レスポンスのコピーに失敗しました: %v", err)
		}
	}
}

// fetchWebPOrigin はWebP用の特別なオリジンフェッチを行います
// webp.go:doWebp関数の移植
func fetchWebPOrigin(ps *ProxyServer, req *http.Request) (*http.Response, error) {
	orgRes, err := ps.FetchFromOrigin(req)
	if err != nil {
		return nil, err
	}

	if orgRes.StatusCode != 200 && orgRes.StatusCode != 304 {
		if err := orgRes.Body.Close(); err != nil {
			log.Printf("レスポンスボディのクローズに失敗しました: %v", err)
		}
		return nil, fmt.Errorf("origin response is not 200 or 304: %s", orgRes.Status)
	}

	return orgRes, nil
}

// handleJPEGToWebPConversion はJPEG→WebP変換処理を行います
func handleJPEGToWebPConversion(w http.ResponseWriter, orgRes *http.Response, ps *ProxyServer) {
	resBytes, err := io.ReadAll(orgRes.Body)
	if err != nil {
		ps.HandleError(w, "Read origin body failed", http.StatusInternalServerError, err)
		return
	}

	// まずWebP変換を試す
	body := io.NopCloser(bytes.NewBuffer(resBytes))
	defer func() {
		if err := body.Close(); err != nil {
			log.Printf("ボディのクローズに失敗しました: %v", err)
		}
	}()
	result, err := convert.ToWebP(body, ps.GetQuality())
	if err == nil {
		// WebP変換成功
		defer result.Data.Reset()
		w.Header().Set("Content-Type", result.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(result.Data.Len()))
		if err := ps.CopyResponse(w, result.Data); err != nil {
			log.Printf("レスポンスのコピーに失敗しました: %v", err)
		}
		return
	}

	// WebP変換失敗時はJPEG変換にフォールバック
	body = io.NopCloser(bytes.NewBuffer(resBytes))
	defer func() {
		if err := body.Close(); err != nil {
			log.Printf("ボディのクローズに失敗しました: %v", err)
		}
	}()
	result, err = convert.ToJPEG(body, ps.GetQuality())
	if err != nil {
		ps.HandleError(w, "Image convert failed", http.StatusInternalServerError, err)
		return
	}
	defer result.Data.Reset()

	w.Header().Set("Content-Type", result.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(result.Data.Len()))
	if err := ps.CopyResponse(w, result.Data); err != nil {
		log.Printf("レスポンスのコピーに失敗しました: %v", err)
	}
}
