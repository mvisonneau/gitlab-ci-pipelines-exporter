package schemas

import (
	"fmt"
	"hash/crc32"
	"strconv"
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
	IncludeSubgroups bool   `yaml:"include_subgroups"`
}

// Wildcards ..
type Wildcards []Wildcard

// WildcardKey ..
type WildcardKey string

// Key ..
func (w Wildcard) Key() WildcardKey {
	return WildcardKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%v", w))))))
}
