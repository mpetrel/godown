package http

import (
	"context"
	"fmt"
	"github.com/mpetrel/godown"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
)

var _ godown.Downloader = (*Downloader)(nil)

type job struct {
	totalBytes int
	canResume  bool
	parts      []*filePart
	originName string
	saveName   string
	fullPath   string
	url        string
}

type filePart struct {
	data []byte
	seq  int
	from int
	to   int
	url  string
}

type Downloader struct {
	maxThread int
	rootDir   string
}

func (d *Downloader) Download(ctx context.Context, url string, savePath string, fileName string) error {
	// 多Goroutine去下载
	jobData, err := d.jobMeta(ctx, url)
	if err != nil {
		return err
	}
	// 根据是否可分片，分发任务执行
	if jobData.canResume {

	}
	return nil
}

func (d *Downloader) jobMeta(ctx context.Context, url string) (*job, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	jobData := &job{url: url}
	// 判断是否支持断点续传
	jobData.canResume = resp.Header.Get("Accept-Ranges") == "bytes"
	// 获取文件大小
	totalBytes, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return nil, err
	}
	jobData.totalBytes = totalBytes
	// 解析文件名
	jobData.originName = parseFileInfoFrom(resp)

	return jobData, nil
}

func (p *filePart) partDown(ctx context.Context) error {
	done := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			done <- err
		}()
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
		if err != nil {
			return
		}
		request.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", p.from, p.to))
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		_, err = resp.Body.Read(p.data)
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func parseFileInfoFrom(resp *http.Response) string {
	contentDisposition := resp.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)

		if err != nil {
			panic(err)
		}
		return params["filename"]
	}
	filename := filepath.Base(resp.Request.URL.Path)
	return filename
}
