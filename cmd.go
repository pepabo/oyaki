package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = ""

var rootCmd = &cobra.Command{
	Use:   "oyaki",
	Short: "Image proxy server with WebP conversion",
	Long: `Oyaki は画像プロキシサーバーで、JPEG画像をWebP形式に変換する機能を提供します。
オリジナル画像サーバーからの画像取得とWebP変換を透過的に行います。`,
	RunE:    runApp,
	Version: getVersion(),
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `アプリケーションのバージョン情報を表示します。`,
	RunE:  showVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Execute はコマンドを実行します
func Execute() error {
	return rootCmd.Execute()
}

// runApp はアプリケーションを実行します
func runApp(cmd *cobra.Command, args []string) error {
	app := NewApp()

	// libvips を初期化
	app.InitializeVips()
	defer app.ShutdownVips()

	// 設定を読み込み
	if err := app.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// pprof サーバーを起動
	app.StartPprofServer()

	// ルーターを設定
	app.SetupRouter()

	// アプリケーションを実行
	return app.Run()
}

// showVersion はバージョン情報を表示します
func showVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("oyaki %s\n", getVersion())
	return nil
}

// getVersion はバージョン情報を取得します
func getVersion() string {
	if version != "" {
		return version
	}

	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "(unknown)"
	}
	return i.Main.Version
}
