package mock

import "github.com/martianmarvin/gidra/client"

type MockHTTPClient struct {
	*MockClient
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		MockClient: &MockClient{},
	}
}

func (c *MockHTTPClient) Page() (*client.Page, error) {
	page := client.NewPage()
	err := page.Parse(c.resp)
	return page, err
}
