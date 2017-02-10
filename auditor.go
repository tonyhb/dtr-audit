package main

import (
	"encoding/base64"
	"fmt"
)

// Auditor implements logic for checking repository access for every user within
// DTR.
//
// There are two types of accounts to check for whilst auditing:
//  - User account repositories
//  - Organization (org) account repositories
//
// User accounts
// =============
//
// These are easy to handle;  if the repository is public then each DTR user
// has read-access while the user who owns the repo has admin access.
// Otherwise, only the user who owns the repo has admin access.
//
// Org accounts
// ============
//
// To audit access to a repository within an organization we need to check if
// the current user is an org admin. If so, they can read/write/edit all org
// repos.
//
// If the user is not an org admin we need to check their access in every team,
// and update their repo permissions according to their team's access.
type Auditor struct {
	// Users is a map Users keyed by user name as a pointer for mutations
	Users map[string]*User `json:"users"`
	// Orgs is a map Orgs keyed by org name as a pointer for mutations
	Orgs map[string]*Org `json:"orgs"`

	// publicRepos represents public repositories found while accessing
	// every user or org's repos.  Storing them here allows us to skip
	// adding them to every user via iterating over the users slice each
	// time.  Instead, we just do it at the end when creating the report.
	publicRepos []Repo

	authHeader string
	host       string
}

func NewAuditor(host, user, pass string) *Auditor {
	encoded := base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", user, pass)),
	)

	return &Auditor{
		Users:      map[string]*User{},
		Orgs:       map[string]*Org{},
		authHeader: fmt.Sprintf("Basic %s", encoded),
		host:       host,
	}
}

func (a *Auditor) Run() error {
	// chain is our pipeline of functions to call which modify internal
	// audit state in order to produce a full repository access audit
	chain := []func() error{
		// first fetch all repositories, giving us a subset of all users
		// and organizations which own at least 1 repository. This is
		// the list of accounts that we need to check
		a.auditAllRepos,

		a.auditOrgs,

		// auditRemainingAccounts fetches accounts that have access to
		// no repos
		a.auditRemainingAccounts,
	}

	for _, f := range chain {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}

func (a *Auditor) auditAllRepos() error {
	fmt.Println("requesting all repos")
	repos, err := a.fetchAllRepos()
	if err != nil {
		return err
	}

	// Now we can iterate through every repository in the system, which
	// gives us all organization names to inspect.
	for _, repo := range repos {
		// create orgs and users within our internal audit state if they
		// do not yet exist
		if repo.AccountType == AccountTypeOrg {
			if _, ok := a.Orgs[repo.AccountName]; !ok {
				a.Orgs[repo.AccountName] = &Org{Name: repo.AccountName}
			}
		} else {

			user, ok := a.Users[repo.AccountName]
			if !ok {
				// create a new user and update our ref
				user = &User{
					Name:  repo.AccountName,
					Repos: map[string]Repo{},
				}
				a.Users[repo.AccountName] = user
			}

			// This repo belongs to the user; assign it to the user
			// with Admin permissions
			repo.Level = Admin
			user.Repos[repo.Name] = repo
		}

		// and if this is public add it to the public slice so
		// we can add this to every user as read-only in the
		// future
		if repo.Visibility == VisibilityPublic {
			a.publicRepos = append(a.publicRepos, repo)
		}
	}

	fmt.Printf(" > found %d repositories\n", len(repos))
	fmt.Printf(" > found %d users\n", len(a.Users))
	fmt.Printf(" > found %d orgs\n", len(a.Orgs))
	return nil
}

func (a *Auditor) auditOrgs() error {
	fmt.Println("auditing all organizations; this may take a while...")
	for orgName, _ := range a.Orgs {
		// fetch the org's teams
		teams, err := a.fetchTeamsForOrg(orgName)
		if err != nil {
			return fmt.Errorf(
				"error fetching teams for org '%s': %s",
				orgName,
				err,
			)
		}

		// i dislike nested fors
		for _, team := range teams {
			if err := a.auditTeam(orgName, team); err != nil {
				return fmt.Errorf(
					"error auditing teams for org '%s': %s",
					orgName,
					err,
				)
			}
		}
	}
	return nil
}

func (a *Auditor) auditTeam(orgName string, team Team) error {
	// get all repos and members for this team, and assign each user the
	// repo permissions granted by the team.
	repos, err := a.fetchReposForTeam(orgName, team.Name)
	if err != nil {
		return err
	}
	users, err := a.fetchMembersForTeam(orgName, team.Name)
	if err != nil {
		return err
	}
	// add each repo to each user. nested loops...
	for _, repo := range repos {
		// set the inner repo struct's Level from the team's
		// Level..
		repo.Repo.Level = repo.Level
		for _, u := range users {
			user, ok := a.Users[u.User.Name]
			// Note: this user may have no repos, therefore wouldn't
			// be populated in the initial repository audit.
			if !ok {
				user = &u.User
				user.Repos = map[string]Repo{}
				a.Users[user.Name] = user

			}
			user.AddRepo(repo.Repo)
		}
	}
	return nil
}

// auditAccounts pulls all users and organizations from the API
func (a *Auditor) auditRemainingAccounts() error {
	fmt.Println("fetching remaining accounts")
	accts, err := a.fetchAccounts()
	if err != nil {
		return err
	}

	// Add each org or user to our internal auditor state
	for _, acc := range accts {
		if acc.IsOrg {
			if _, ok := a.Orgs[acc.Name]; !ok {
				a.Orgs[acc.Name] = &Org{
					ID:   acc.ID,
					Name: acc.Name,
				}
			}
		} else {
			if _, ok := a.Users[acc.Name]; !ok {
				fmt.Println(" > found new user with no repos")
				a.Users[acc.Name] = &User{
					ID:      acc.ID,
					Name:    acc.Name,
					IsAdmin: acc.IsAdmin,
				}
			} else {
				a.Users[acc.Name].IsAdmin = acc.IsAdmin
				a.Users[acc.Name].ID = acc.ID
			}
		}
	}
	return nil
}
