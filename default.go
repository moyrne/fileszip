package fileszip

import (
	"io"
	"strings"
)

type DefaultUserHook struct {
	replacer *strings.Replacer
}

func (d *DefaultUserHook) TransPath(p Sources) string {
	return d.replacer.Replace(p.Url)
}

var DefaultFilesZip = NewFilesZip()

func AsyncRead(params []Sources) io.Reader {
	return DefaultFilesZip.ASyncRead(params)
}

func WriteFile(params []Sources, writer io.Writer) (err error) {
	return DefaultFilesZip.WriteFile(params, writer)
}
