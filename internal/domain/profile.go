package domain

// DefaultProfileName is the default profile name.
const DefaultProfileName = "default"

// Profile bundles a catalog under a named profile.
type Profile struct {
	Name    string
	Catalog Catalog
}

// ProfileStore contains profiles and caller mappings.
type ProfileStore struct {
	Profiles map[string]Profile
	Callers  map[string]string
}
