package wrpparser

import (
	"github.com/stretchr/testify/mock"
	"github.com/xmidt-org/wrp-go/v3"
)

type MockDeviceFinder struct {
	mock.Mock
}

func (f *MockDeviceFinder) FindDeviceID(msg *wrp.Message) (string, error) {
	args := f.Called(msg)
	return args.String(0), args.Error(1)
}

type MockClassifier struct {
	mock.Mock
}

func (c *MockClassifier) Label(msg *wrp.Message) (string, bool) {
	args := c.Called(msg)
	return args.String(0), args.Bool(1)
}
