package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	localRatelimit "go.uber.org/ratelimit"
)

func TestNewLocalLimiter(t *testing.T) {
	assert.Equal(t, Local{localRatelimit.New(10)}, NewLocalLimiter(10))
}
