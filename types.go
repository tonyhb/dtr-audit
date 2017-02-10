package main

const (
	AccountTypeOrg  = "organization"
	AccountTypeUser = "user"

	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// Repo represents an individual repository, including access level and
// visibility (public or private)
type Repo struct {
	// Name represents the full name of the repository, ie "foo/bar".
	ID          string `json:"id"`
	Name        string `json:"name"`
	Visibility  string `json:"visibility"`
	Level       Access `json:"access"`
	AccountName string `json:"namespace"`
	AccountType string `json:"namespaceType"`
}

// TeamRepo is the type returned when querying for the repositories that a
// team can access.  Note that the API returns the access level as a top
// level field - not in the repo. :(
type TeamRepo struct {
	Repo  Repo   `json:"repository"`
	Level Access `json:"accessLevel"`
}

// Account represents a single account fetched from the list of accounts
// via the API
type Account struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsOrg   bool   `json:"isOrg"`
	IsAdmin bool   `json:"isAdmin"`
}

// Team represents a team within an organization, including all users which
// are a member of the team and all repositories it has access to.
type Team struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
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
	ID      string `json:"id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"isAdmin"`
	// Repos is a map of repository names to repositories they have access
	// to.  It's a map to ensure uniqueness of repos.
	// Note that this holds **all** repositories the user has access to
	// for the audit log, not just the ones they own
	Repos map[string]Repo
}

func (u *User) AddRepo(r Repo) {
	if existing, ok := u.Repos[r.Name]; ok {
		// our existing repo has greater permissions; quit
		if existing.Level > r.Level {
			return
		}
	}
	u.Repos[r.Name] = r
}

type RepoWrapper struct {
	Repos []Repo `json:"repositories"`
}

type TeamWrapper struct {
	Teams []Team `json:"teams"`
}

type TeamRepoWrapper struct {
	Repos []TeamRepo `json:"repositoryAccessList"`
}

type TeamMemberWrapper struct {
	Members []TeamMember `json:"members"`
}

type TeamMember struct {
	User User `json:"member"`
}
