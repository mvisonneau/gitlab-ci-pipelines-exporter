package schemas

import (
	"hash/crc32"
	"strconv"
)

// Environment ..
type Environment struct {
	ProjectPathWithNamespace string
	ID                       uint64

	Name             string
	ExternalURL      string
	LatestDeployment Deployment
}

// EnvironmentKey ..
type EnvironmentKey string

// Key ..
func (e Environment) Key() EnvironmentKey {
	return EnvironmentKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(e.ProjectPathWithNamespace + strconv.Itoa(int(e.ID)))))))
}

// Environments allows us to keep track of all the Environment
// objects we have discovered
type Environments map[EnvironmentKey]Environment

// Count returns the amount of environments in the map
func (envs Environments) Count() int {
	return len(envs)
}
