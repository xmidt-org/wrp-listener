package store

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImplementsInterfaces(t *testing.T) {
	var (
		inmem interface{}
	)
	assert := assert.New(t)
	inmem = CreateInMemStore()
	_, ok := inmem.(Hook)
	assert.True(ok, "not an webhook Hook")
	_, ok = inmem.(Listener)
	assert.True(ok, "not an webhook Listener")
}
