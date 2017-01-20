package datasource

import (
	"bufio"
	"bytes"
	"io"
	"mime"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/martianmarvin/gidra/client/httpclient"
)

// Finds extension from MIME type. Returns empty string if none found.
func mimeExts(mt string) ([]string, error) {
	exts, err := mime.ExtensionsByType(mt)
	if err != nil {
		return nil, err
	}
	if len(exts) == 0 {
		return nil, ErrUnsupportedType
	}
	return exts, nil
}

// ReaderFor returns a reader that supports the specified file extension or MIME
// type
func ReaderFor(format string) (ReadableTable, error) {
	var err error
	reader, err := NewReader(strings.TrimLeft(format, "."))
	if err == nil {
		return reader, nil
	}

	// Check mime type
	exts, err := mimeExts(format)
	if err != nil {
		return nil, err
	}

	for _, ext := range exts {
		reader, err := NewReader(strings.TrimLeft(ext, "."))
		if err == nil {
			return reader, nil
		}
	}
	//TODO: Custom error for unsupported type, showing the error
	return nil, ErrUnsupportedType
}

func fromFile(fp, format string) (ReadableTable, error) {
	var err error
	reader, err := ReaderFor(format)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	_, err = reader.ReadFrom(f)
	return reader, err
}

// FromFile returns a reader reading from the specified file
func FromFile(fp string) (ReadableTable, error) {
	return FromFileType(fp, "")
}

// FromFileType returns a reader that can read the specified format. The
// format can be a file extension or a MIME type
func FromFileType(fp string, format string) (ReadableTable, error) {
	if len(format) == 0 {
		ext := path.Ext(fp)
		if len(ext) == 0 {
			return nil, ErrUnsupportedType
		}
		return fromFile(fp, ext)
	}
	return fromFile(fp, format)
}

// TODO Parse content-type header and extension to determine mime type
func fromURL(u *url.URL) (ReadableTable, error) {
	return nil, nil
}

// FromURL returns a reader reading from the specified URL. The address can also
// be a local file with the file:// scheme
func FromURL(u *url.URL) (ReadableTable, error) {
	if u.Scheme == "file" || u.Scheme == "" {
		fp := u.Host + u.Path
		return FromFile(fp)
	} else {
		return fromURL(u)
	}
}

// ReadLines is a helper that reads lines into a string slice from a local or
// remote resource
func ReadLines(path string) ([]string, error) {
	var lines []string
	var err error
	var f io.Reader

	// Parse to url to determine local or remote
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if len(u.Host) < 2 {
		// Must just have path, and be a local file
		f, err = os.Open(path)
		if err != nil {
			return nil, err
		}
	} else {
		body, err := httpclient.Get(path)
		if err != nil {
			return nil, err
		}
		f = bytes.NewReader(body)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
