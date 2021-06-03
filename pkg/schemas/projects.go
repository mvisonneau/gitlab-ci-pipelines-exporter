package schemas

import (
	"hash/crc32"
	"strconv"

	"github.com/mvisonneau/gitlab-ci-pipelines-exporter/pkg/config"
)

// Project ..
type Project struct {
	config.Project

	Topics string
}

// ProjectKey ..
type ProjectKey string

// Projects ..
type Projects map[ProjectKey]Project

// Key ..
func (p Project) Key() ProjectKey {
	return ProjectKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(p.Name)))))
}

// NewProject ..
func NewProject(name string) Project {
	return Project{Project: config.NewProject(name)}
}
