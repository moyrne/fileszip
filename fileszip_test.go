package fileszip

import (
	"archive/zip"
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestWriteFile(t *testing.T) {
	src := "./tests"
	dst := "./testsout"
	os.RemoveAll(src)
	if err := os.MkdirAll(src, 0775); err != nil {
		panic(err)
	}
	os.RemoveAll(dst)
	if err := os.MkdirAll(dst, 0775); err != nil {
		panic(err)
	}

	type args struct {
		repeat int
	}

	tests := make([]struct {
		name string
		args args
		want string
	}, 100)

	o := sha512.New()
	o.Write([]byte("test repeat"))
	repeatValue := hex.EncodeToString(o.Sum(nil))

	sources := make([]Sources, len(tests))
	for i := 0; i < len(tests); i++ {
		tests[i].name = "repeat-" + strconv.Itoa(i+1)
		tests[i].args.repeat = i + 1
		filename := path.Join(src, tests[i].name)
		var err error
		tests[i].want, err = writeTestFile(filename, repeatValue, i+1)
		if err != nil {
			t.Error(err)
			return
		}
		sources[i] = Sources{Url: filename}
	}

	zipFilename := path.Join(dst, "file.zip")
	writeFile(zipFilename, AsyncRead(sources))

	reader, err := zip.OpenReader(zipFilename)
	if err != nil {
		t.Error(err)
		return
	}

	for i, source := range sources {
		if err := func() error {
			f, err := reader.Open(DefaultFilesZip.userHook.TransPath(source))
			if err != nil {
				return errors.Wrap(err, source.Url)
			}
			defer f.Close()

			o := md5.New()
			if _, err := io.Copy(o, f); err != nil {
				return errors.Wrap(err, source.Url)
			}

			oSum := hex.EncodeToString(o.Sum(nil))
			if tests[i].want != oSum {
				return errors.Wrapf(errors.New("want and file is not equal"), "i: %d, url: %s, want: %s, sum: %s", i, source.Url, tests[i].want, oSum)
			}

			return nil
		}(); err != nil {
			t.Error(err)
			return
		}
	}
}

func writeTestFile(filename string, value string, repeat int) (md5Hash string, err error) {
	f, err := os.Create(filename)
	if err != nil {
		return "", errors.Wrap(err, filename)
	}

	defer f.Close()
	o := md5.New()

	byteValue := []byte(value)
	for i := 0; i < repeat; i++ {
		if _, err := f.Write(byteValue); err != nil {
			return "", errors.Wrap(err, filename)
		}
		o.Write(byteValue)
	}

	return hex.EncodeToString(o.Sum(nil)), nil
}

func BenchmarkWriteFile(b *testing.B) {
	src := "./tests"
	dst := "./testsout"
	os.RemoveAll(src)
	if err := os.MkdirAll(src, 0775); err != nil {
		panic(err)
	}
	os.RemoveAll(dst)
	if err := os.MkdirAll(dst, 0775); err != nil {
		panic(err)
	}

	type args struct {
		repeat int
	}

	tests := make([]struct {
		name string
		args args
		want string
	}, 100)

	o := sha512.New()
	o.Write([]byte("test repeat"))
	repeatValue := hex.EncodeToString(o.Sum(nil))

	sources := make([]Sources, len(tests))
	for i := 0; i < len(tests); i++ {
		tests[i].name = "repeat-" + strconv.Itoa(i+1)
		tests[i].args.repeat = i + 1
		filename := path.Join(src, tests[i].name)
		var err error
		tests[i].want, err = writeTestFile(filename, repeatValue, i+1)
		if err != nil {
			b.Error(err)
			return
		}
		sources[i] = Sources{Url: filename}
	}

	zipFilename := path.Join(dst, "file.zip")

	b.Run("zipfile", func(b *testing.B) {
		b.ResetTimer()
		writeFile(zipFilename, AsyncRead(sources))
		b.StopTimer()
	})

	reader, err := zip.OpenReader(zipFilename)
	if err != nil {
		b.Error(err)
		return
	}

	for i, source := range sources {
		if err := func() error {
			f, err := reader.Open(DefaultFilesZip.userHook.TransPath(source))
			if err != nil {
				return errors.Wrap(err, source.Url)
			}
			defer f.Close()

			o := md5.New()
			if _, err := io.Copy(o, f); err != nil {
				return errors.Wrap(err, source.Url)
			}

			oSum := hex.EncodeToString(o.Sum(nil))
			if tests[i].want != oSum {
				return errors.Wrapf(errors.New("want and file is not equal"), "i: %d, url: %s, want: %s, sum: %s", i, source.Url, tests[i].want, oSum)
			}

			return nil
		}(); err != nil {
			b.Error(err)
			return
		}
	}
}

func writeFile(filename string, r io.Reader) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		panic(err)
	}
}
