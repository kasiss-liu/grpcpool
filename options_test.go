package pool

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	opt, err := NewOptions(10, []string{"127.0.0.1:8899"})
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 10, opt.Cap)
	assert.Contains(t, opt.Targets, "127.0.0.1:8899")
	assert.Equal(t, "127.0.0.1:8899", opt.getTarget())
	assert.Equal(t, 5*time.Second, opt.IdleTimeout)
}

func TestOptionsValid(t *testing.T) {
	opt := Options{}

	err := opt.validate()
	assert.IsType(t, err, ErrTargetEmpty)

	opt.Targets = []string{"127.0.0.1:8899"}

	err = opt.validate()
	assert.IsType(t, err, ErrOptionValid)
}
