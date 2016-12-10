package client

//HTTPClient performs http requests
type HTTPClient interface {
	Client

	// Page returns the last response parsed into a *Page
	Page() (*Page, error)
}
