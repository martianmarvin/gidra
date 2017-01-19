package datasource

import "io"

//NopWriter is a dummy writer that silently discards all data written to it

type NopWriter struct{}

func (w *NopWriter) WriteTo(writer io.Writer) (n int64, err error) {
	return 0, nil
}

func (w *NopWriter) SetColumns(cols []string) error {
	return nil
}

func (w *NopWriter) Append(row *Row) error {
	return nil
}

func (w *NopWriter) Close() error {
	return nil
}

func (w *NopWriter) Filter(fn FilterFunc) error {
	return nil
}
