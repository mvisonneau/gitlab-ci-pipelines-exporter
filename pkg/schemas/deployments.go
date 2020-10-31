package schemas

import "time"

// Deployment ..
type Deployment struct {
	RefKind       RefKind
	RefName       string
	AuthorEmail   string
	CreatedAt     time.Time
	Duration      time.Duration
	CommitShortID string
	Status        string
}
