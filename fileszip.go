package fileszip

import (
	"archive/zip"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type DefaultUserHook struct {
	replacer *strings.Replacer
}

func (d *DefaultUserHook) TransPath(s string) string {
	return d.replacer.Replace(s)
}

var DefaultFilesZip = &FilesZip{
	client: &http.Client{
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 10 * time.Second}).DialContext,
			IdleConnTimeout:       10 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
		},
		Timeout: 10 * time.Second,
	},
	userHook: &DefaultUserHook{
		replacer: strings.NewReplacer("https://", "", "http://", "", "/", "-", ":", "-"),
	},
}

func WriteFile(paths []string, writer io.Writer) (err error) {
	return DefaultFilesZip.WriteFile(paths, writer)
}

type (
	Client interface {
		Get(url string) (*http.Response, error)
	}
	UserHook interface {
		TransPath(s string) string
	}
)

type FilesZip struct {
	client Client

	userHook UserHook
}

func (f *FilesZip) WriteFile(paths []string, writer io.Writer) (err error) {
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	for _, path := range paths {
		// TODO 可能需要考虑往外抛执行状态

		log.Println("start get:", path)

		resp, err := f.client.Get(path)
		if err != nil {
			return errors.Wrapf(err, "get file failed: %s", path)
		}
		// 自定义文件名
		pathWriter, err := zipWriter.Create(f.userHook.TransPath(path))
		if err != nil {
			return errors.Wrapf(err, "create zip file failed: %s", path)
		}
		// 可能有超时的问题
		if _, err := io.Copy(pathWriter, resp.Body); err != nil {
			return errors.Wrap(err, "copy body to path writer failed")
		}
	}

	return nil
}
