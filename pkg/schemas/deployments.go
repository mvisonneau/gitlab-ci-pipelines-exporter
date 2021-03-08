package schemas

// Deployment ..
type Deployment struct {
	JobID           int
	RefKind         RefKind
	RefName         string
	Username        string
	Timestamp       float64
	DurationSeconds float64
	CommitShortID   string
	Status          string
}
