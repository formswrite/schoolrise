package domain

type HierarchyLevel string

const (
	LevelNational   HierarchyLevel = "national"
	LevelRegion     HierarchyLevel = "region"
	LevelPrefecture HierarchyLevel = "prefecture"
	LevelDelegation HierarchyLevel = "delegation"
	LevelInstitution HierarchyLevel = "institution"
	LevelClass      HierarchyLevel = "class"
	LevelGroup      HierarchyLevel = "group"
	LevelStudent    HierarchyLevel = "student"
)

type Subject string

const (
	SubjectFrench Subject = "french"
	SubjectMaths  Subject = "maths"
)

type Period struct {
	ID        int64
	Code      string
	Label     string
	StartDate string
	EndDate   string
}

type Scale struct {
	Code  string
	Label string
}

type ScaleBand struct {
	ScaleCode string
	Ordinal   int
	Code      string
	Label     string
}

type RoleScope struct {
	Role    string
	ScopeID int64
	Level   HierarchyLevel
}
