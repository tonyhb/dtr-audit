package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// url is a utility functionm for appending a URL to the endpoint we need to hit
func (a *Auditor) url(endpoint string) string {
	return fmt.Sprintf("%s/%s", a.host, endpoint)
}

func (a *Auditor) get(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		a.url(endpoint),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}
	resp, err := a.do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %s", err)
	}
	return resp, nil
}

func (a *Auditor) fetchAllRepos() ([]Repo, error) {
	resp, err := a.get("api/v0/repositories?limit=999999")
	if err != nil {
		return nil, fmt.Errorf("error requesting all repos: %s", err)
	}
	defer resp.Body.Close()

	aux := RepoWrapper{}
	if err := json.NewDecoder(resp.Body).Decode(&aux); err != nil {
		return nil, fmt.Errorf("error decoding accounts: %s", err)
	}
	return aux.Repos, nil
}

// fetchAccounts pulls all users and organizations from the API
func (a *Auditor) fetchAccounts() ([]Account, error) {
	fmt.Println("fetching accounts")

	// Make a request to fetch all accounts
	resp, err := a.get("enzi/v0/accounts")
	if err != nil {
		return nil, fmt.Errorf("error requesting accounts: %s", err)
	}
	defer resp.Body.Close()

	// Note that the API response for the list of accounts is wrapped in
	// an object.
	aux := struct {
		Accounts []Account `json:"accounts"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&aux); err != nil {
		return nil, fmt.Errorf("error decoding accounts: %s", err)
	}

	return aux.Accounts, nil
}

// do adds authorization headers to an http.Request, makes the request and
// returns the response.
//
// This uses retry defined in util.go to attempt the http request up to 3 times
// before finally packing up and going home to its fam.
func (a *Auditor) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", a.authHeader)

	// create a response which is caught inside the closure below which
	// performs the http request.
	//
	// this allows us to pass the http requesting function into retry()
	// whilst capturing the response outside of the function
	var resp *http.Response

	// closure alert
	f := func() error {
		var err error
		if resp, err = client.Do(req); err != nil {
			return err
		}

		// validate our http codes to make sure things are as we expected
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf(
				"unexpected status code requesting uri %s: %d",
				req.URL.String(),
				resp.StatusCode,
			)
		}

		// yay
		return nil
	}

	return resp, retry(f)
}

// fetchTeamsForOrg returns all repositories owned by a user or
// organization
func (a *Auditor) fetchTeamsForOrg(name string) ([]Team, error) {
	resp, err := a.get(fmt.Sprintf("enzi/v0/accounts/%s/teams?limit=5000", name))
	if err != nil {
		return nil, fmt.Errorf("error requesting teams: %s", err)
	}
	defer resp.Body.Close()
	aux := TeamWrapper{}
	if err := json.NewDecoder(resp.Body).Decode(&aux); err != nil {
		return nil, fmt.Errorf("error decoding teams: %s", err)
	}
	return aux.Teams, nil
}

// fetchReposForTeam returns all repositories a team has access to
// organization
func (a *Auditor) fetchReposForTeam(orgName, teamName string) ([]TeamRepo, error) {
	resp, err := a.get(fmt.Sprintf(
		"api/v0/accounts/%s/teams/%s/repositoryAccess?limit=100000",
		orgName,
		teamName,
	))
	if err != nil {
		return nil, fmt.Errorf("error requesting repos for team: %s", err)
	}
	defer resp.Body.Close()
	aux := TeamRepoWrapper{}
	if err := json.NewDecoder(resp.Body).Decode(&aux); err != nil {
		return nil, fmt.Errorf("error decoding repos for team: %s", err)
	}
	return aux.Repos, nil
}

func (a *Auditor) fetchMembersForTeam(orgName, teamName string) ([]TeamMember, error) {
	resp, err := a.get(fmt.Sprintf(
		"enzi/v0/accounts/%s/teams/%s/members?limit=10000",
		orgName,
		teamName,
	))
	if err != nil {
		return nil, fmt.Errorf("error requesting users for team: %s", err)
	}
	defer resp.Body.Close()
	aux := TeamMemberWrapper{}
	if err := json.NewDecoder(resp.Body).Decode(&aux); err != nil {
		return nil, fmt.Errorf("error decoding users for team: %s", err)
	}
	return aux.Members, nil
}
