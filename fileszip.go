package fileszip

import (
	"archive/zip"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"log"
	"net/http"
)

type (
	Client interface {
		Get(url string) (*http.Response, error)
	}
	UserHook interface {
		TransPath(p Sources) string
	}
)

type FilesZip struct {
	debug    bool
	client   Client
	userHook UserHook
}

type Sources struct {
	Url   string      `json:"url"`
	Extra interface{} `json:"extra"`
}

func (s Sources) String() string {
	data, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func (f *FilesZip) ASyncRead(sources []Sources) io.Reader {
	r, w := io.Pipe()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("async read goroutine recovered", r)
			}
		}()
		defer w.Close()
		if err := WriteFile(sources, w); err != nil {
			if err := w.CloseWithError(err); err != nil {
				log.Println("close pipe failed", err.Error())
			}
			return
		}
	}()

	return r
}

func (f *FilesZip) WriteFile(sources []Sources, writer io.Writer) (err error) {
	zipWriter := zip.NewWriter(writer)
	defer zipWriter.Close()

	for _, source := range sources {
		// TODO 可能需要考虑往外抛执行状态

		by, err := json.Marshal(source)
		if err != nil {
			return errors.WithStack(err)
		}
		if f.debug {
			log.Println("start get:", string(by))
		}
		if err := f.downloadFile(zipWriter, source); err != nil {
			return err
		}
	}

	return nil
}

func (f *FilesZip) downloadFile(zipWriter *zip.Writer, sources Sources) error {
	resp, err := f.client.Get(sources.Url)
	if err != nil {
		return errors.Wrapf(err, "get file failed: %s", sources)
	}
	defer resp.Body.Close()

	// 自定义文件名
	pathWriter, err := zipWriter.Create(f.userHook.TransPath(sources))
	if err != nil {
		return errors.Wrapf(err, "create zip file failed: %s", sources)
	}
	// 可能有超时的问题
	if _, err := io.Copy(pathWriter, resp.Body); err != nil {
		return errors.Wrap(err, "copy body to path writer failed")
	}

	return nil
}
