package schemas

// Wildcard is a specific handler to dynamically search projects
type Wildcard struct {
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
