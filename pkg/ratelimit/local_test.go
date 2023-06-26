package ratelimit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocalLimiter(t *testing.T) {
	assert.IsType(t, Local{}, NewLocalLimiter(10, 1))
}
