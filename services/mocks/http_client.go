package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock of HTTPClient.
type MockHTTPClient struct {
	mock.Mock
}

// Do is a mock of the Do method.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
