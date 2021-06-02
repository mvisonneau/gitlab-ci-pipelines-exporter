package config

import (
	"fmt"
	"hash/crc32"
	"strconv"

	"github.com/creasty/defaults"
)

// Wildcard is a specific handler to dynamically search projects
type Wildcard struct {
	// ProjectParameters holds parameters specific to the projects which
	// will be discovered using this wildcard
	ProjectParameters `yaml:",inline"`

	Search   string        `yaml:"search"`
	Owner    WildcardOwner `yaml:"owner"`
	Archived bool          `yaml:"archived"`
}

// WildcardOwner ..
type WildcardOwner struct {
	Name             string `yaml:"name"`
	Kind             string `yaml:"kind"`
	IncludeSubgroups bool   `default:"false" yaml:"include_subgroups"`
}

// Wildcards ..
type Wildcards []Wildcard

// WildcardKey ..
type WildcardKey string

// Key ..
func (w Wildcard) Key() WildcardKey {
	return WildcardKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%v", w))))))
}

// NewWildcard returns a new wildcard with the default parameters
func NewWildcard() (w Wildcard) {
	defaults.MustSet(&w)
	return
}
