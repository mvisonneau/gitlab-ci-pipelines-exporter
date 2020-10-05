package schemas

// Wildcard is a specific handler to dynamically search projects
type Wildcard struct {
	Parameters `yaml:",inline"`

	Search string `yaml:"search"`
	Owner  struct {
		Name             string `yaml:"name"`
		Kind             string `yaml:"kind"`
		IncludeSubgroups bool   `yaml:"include_subgroups"`
	} `yaml:"owner"`
	Archived bool `yaml:"archived"`
}

// Wildcards ..
type Wildcards []Wildcard
