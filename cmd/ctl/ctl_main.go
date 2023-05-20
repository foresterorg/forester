package main

import (
	"context"
	"forester/internal/config"
	"forester/internal/img"
	"forester/internal/log"
	"os"
)

func main() {
	log.Initialize()
	err := config.Initialize()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	file, err := os.Open("f37-truncated.iso")
	if err != nil {
		panic(err)
	}
	err = img.UploadImage(ctx, 1, file)
	if err != nil {
		panic(err)
	}
}
