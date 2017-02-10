package main

const (
	Public  = "public"
	Private = "private"

	Admin = "admin"
)

// Repo represents an individual repository within a team or user,
// including access level and visibility (public or private)
type Repo struct {
	// Name represents the full name of the repository, ie "foo/bar".
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
	Access     string `json:"access"`
}

// Account represents a single account fetched from the list of accounts
// via the API
type Account struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	IsOrg bool   `json:"isOrg"`
}

// Team represents a team within an organization, including all users which
// are a member of the team and all repositories it has access to.
type Team struct {
	ID    string
	Users []Account
	Repos []Repo
}

// Org represents an organization within DTR. An org has many repos which are
// visibile to org admins and to normal users via teams.
type Org struct {
	ID    string
	Name  string
	Teams []Team
}

// User represents an individual user within DTR with the repositories they
// have access to, including their access level
type User struct {
	ID   string
	Name string
	// Repos is a map of repository names to repositories they have access
	// to.  It's a map to ensure uniqueness of repos.
	// Note that this holds **all** repositories the user has access to
	// for the audit log, not just the ones they own
	Repos map[string]Repo
}
