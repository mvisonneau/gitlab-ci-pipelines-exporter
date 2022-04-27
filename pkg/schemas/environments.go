package schemas

import (
	"hash/crc32"
	"strconv"
)

// Environment ..
type Environment struct {
	ProjectName      string
	ID               int
	Name             string
	ExternalURL      string
	Available        bool
	LatestDeployment Deployment

	OutputSparseStatusMetrics bool
}

// EnvironmentKey ..
type EnvironmentKey string

// Key ..
func (e Environment) Key() EnvironmentKey {
	return EnvironmentKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(e.ProjectName + e.Name)))))
}

// Environments allows us to keep track of all the Environment objects we have discovered.
type Environments map[EnvironmentKey]Environment

// Count returns the amount of environments in the map.
func (envs Environments) Count() int {
	return len(envs)
}

// DefaultLabelsValues ..
func (e Environment) DefaultLabelsValues() map[string]string {
	return map[string]string{
		"project":     e.ProjectName,
		"environment": e.Name,
	}
}

// InformationLabelsValues ..
func (e Environment) InformationLabelsValues() (v map[string]string) {
	v = e.DefaultLabelsValues()
	v["environment_id"] = strconv.Itoa(e.ID)
	v["external_url"] = e.ExternalURL
	v["kind"] = string(e.LatestDeployment.RefKind)
	v["ref"] = e.LatestDeployment.RefName
	v["current_commit_short_id"] = e.LatestDeployment.CommitShortID
	v["latest_commit_short_id"] = ""
	v["available"] = strconv.FormatBool(e.Available)
	v["username"] = e.LatestDeployment.Username

	return
}
