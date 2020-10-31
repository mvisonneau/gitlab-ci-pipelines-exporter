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
	TagsRegexp       string
}

// EnvironmentKey ..
type EnvironmentKey string

// Key ..
func (e Environment) Key() EnvironmentKey {
	return EnvironmentKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(e.ProjectName + strconv.Itoa(int(e.ID)))))))
}

// Environments allows us to keep track of all the Environment
// objects we have discovered
type Environments map[EnvironmentKey]Environment

// Count returns the amount of environments in the map
func (envs Environments) Count() int {
	return len(envs)
}

// DefaultLabelsValues ..
func (e Environment) DefaultLabelsValues() map[string]string {
	return map[string]string{
		// TODO: Replace with the pathwithnamespace..
		"project":     e.ProjectName,
		"environment": e.Name,
	}
}

// InformationLabelsValues ..
func (e Environment) InformationLabelsValues() (v map[string]string) {
	v = e.DefaultLabelsValues()
	v["external_url"] = e.ExternalURL
	v["kind"] = string(e.LatestDeployment.RefKind)
	v["ref"] = e.LatestDeployment.RefName
	v["current_commit_short_id"] = e.LatestDeployment.CommitShortID
	v["available"] = strconv.FormatBool(e.Available)
	v["author_email"] = e.LatestDeployment.AuthorEmail

	return
}
