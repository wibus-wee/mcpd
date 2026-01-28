//go:build !test

package main

import "embed"

//go:embed all:frontend/dist
var Assets embed.FS
