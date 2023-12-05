package godown

import "context"

type Downloader interface {
	Download(ctx context.Context, url string, savePath string, fileName string) error
}
