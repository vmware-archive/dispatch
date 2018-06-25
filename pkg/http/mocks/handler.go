package mocks

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

// HandlerMock mocks http.Handler
type HandlerMock struct {
	mock.Mock
}

// ServeHTTP implements http.Handler interface
func (_m *HandlerMock) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	_m.Called(rw, req)

}
