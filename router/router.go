package router

import (
	"net/http"
	"path/filepath"
)

// Router は汎用HTTPルーターのインターフェースです
type Router interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// ExtensionRouter は拡張子に基づいてハンドラーを振り分けるルーターです
type ExtensionRouter struct {
	handlers       map[string]http.HandlerFunc
	defaultHandler http.HandlerFunc
}

// NewExtensionRouter は新しいExtensionRouterを作成します
func NewExtensionRouter() *ExtensionRouter {
	return &ExtensionRouter{
		handlers: make(map[string]http.HandlerFunc),
	}
}

// HandleFunc はパターン（拡張子）に対応するハンドラー関数を登録します
// pattern が空文字列の場合はデフォルトハンドラーとして扱います
func (er *ExtensionRouter) HandleFunc(pattern string, handler http.HandlerFunc) {
	if pattern == "" {
		er.defaultHandler = handler
	} else {
		er.handlers[pattern] = handler
	}
}

// ServeHTTP は http.Handler インターフェースを実装します
func (er *ExtensionRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// パスから拡張子を取得
	ext := filepath.Ext(r.URL.Path)

	// 拡張子に対応するハンドラーを検索
	if handler, exists := er.handlers[ext]; exists {
		handler(w, r)
		return
	}

	// デフォルトハンドラーがあれば使用
	if er.defaultHandler != nil {
		er.defaultHandler(w, r)
		return
	}

	// どのハンドラーも見つからない場合は404
	http.NotFound(w, r)
}
