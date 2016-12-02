package datasource

import (
	"mime"
	"net/url"
	"os"
	"path"
	"strings"
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

func fromFile(fp, format string) (ReadableTable, error) {
	var err error
	reader, err := NewReader(strings.TrimLeft(format, "."))
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
	ext := path.Ext(fp)
	if len(ext) == 0 {
		return nil, ErrUnsupportedType
	}
	return fromFile(fp, ext)
}

// TODO Parse content-type header and extension to determine mime type
func fromURL(u *url.URL) (ReadableTable, error) {
	return nil, nil
}

// FromURL returns a reader reading from the specified URL. The address can also
// be a local file with the file:// scheme
func FromURL(u *url.URL) (ReadableTable, error) {
	if u.Scheme == "file" {
		fp := u.Host + u.Path
		return FromFile(fp)
	} else {
		return fromURL(u)
	}
}
