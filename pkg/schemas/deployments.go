package schemas

import "time"

// Deployment ..
type Deployment struct {
	RefKind   RefKind
	RefName   string
	Author    string
	Time      time.Time
	Duration  time.Duration
	CommitSHA string
	Status    string
}
