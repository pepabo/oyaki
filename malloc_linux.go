//go:build linux

package main

/*
#cgo pkg-config: vips
#include <malloc.h>
#include <vips/vips.h>
*/
import "C"

// mallocTrim は glibc に対してフリーなメモリを OS に返すよう要求する。
// libvips (CGo) の処理後に蓄積する malloc アリーナを縮小するために使う。
func mallocTrim() {
	C.malloc_trim(0)
}

func setVipsConcurrency(n int) {
	C.vips_concurrency_set(C.int(n))
}
