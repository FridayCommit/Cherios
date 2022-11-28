package sonarqube

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
)

const (
	ApplicationsQualifier = "APP" // Used for search
	PortfolioQualifier    = "VW"  // Used for search
	PortfoliosQualifier   = "SVW" // Used for search
	ProjectQualifier      = "TRK" // Used for search
)

var (
	SonarUrl = "https://sonarqube.snowdev.io"
)

// TODO
// We need to get or pass the default branch from the github call
// TODO ADD ERROR HANDLING FOR EVERYTHING
type searchResp struct {
	Paging struct {
		PageIndex int `json:"pageIndex"`
		PageSize  int `json:"pageSize"`
		Total     int `json:"total"`
	} `json:"paging"`
	Components []struct {
		Key       string `json:"key"`
		Name      string `json:"name"`
		Qualifier string `json:"qualifier"`
		Project   string `json:"project"`
	} `json:"components"`
}
type createResp struct {
	Project struct {
		Key        string `json:"key"`
		Name       string `json:"name"`
		Qualifier  string `json:"qualifier"`
		Visibility string `json:"visibility"`
	} `json:"project"`
}

// sonarqubeCall calls the authenticates and calls the sonarqube webApi
func sonarqubeCall(method string, url string, form url.Values, contentType string) (*http.Response, error) {
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod it's just for good development
	if err != nil {
		return nil, fmt.Errorf("cannot find SonarQube Token")
	}
	token := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	req, err := http.NewRequest(method, url, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(token, "")
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// SearchSonarQube Checks if the project already exists in sonarqube. Returns False with error
func SearchSonarQube(qualifier string, search string) (bool, error) {
	apiUrl, err := url.Parse("/api/components/search") // TODO this is a little ugly but it works super well. maybe works with forms
	if err != nil {
		return false, err
	}
	values := apiUrl.Query()
	values.Add("qualifiers", qualifier) // enum maybe? maybe not needed
	values.Add("q", search)
	apiUrl.RawQuery = values.Encode()
	resp, err := sonarqubeCall(http.MethodGet, SonarUrl+apiUrl.String(), nil, "application/json; charset=UTF-8")
	defer resp.Body.Close()
	if err != nil {
		return false, err
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
		return false, fmt.Errorf("SearchSonarQube Http Response: %v", resp.StatusCode)
	}
	var result searchResp
	body, err := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &result); err != nil {
		return false, err
	}
	// The sonarqube API searches partially. So if someone would name their repo devops-app then devops-applications would also show u<p.
	for _, component := range result.Components { // TODO we could break this out to its own function i guess ?
		fields := reflect.ValueOf(component)
		for i := 0; i < fields.NumField(); i++ {
			if fields.Field(i).String() == search {
				return true, nil
			}
		}
	}
	return false, nil
}

// createProject Creates a SonarQube Project via WebAPI
func createProject(repositoryPayload github.RepositoryPayload) error { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	apiUrl := "/api/projects/create"
	form := url.Values{}
	form.Add("name", repositoryPayload.Repository.Name)
	form.Add("project", repositoryPayload.Repository.Name)
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiUrl, form, "application/x-www-form-urlencoded")
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	var result createResp
	body, err := io.ReadAll(resp.Body)                   // response body is []byte
	if err = json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		return err
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
		return fmt.Errorf("createProject Http Response: %v", resp.StatusCode)
	}
	return nil
}

// createPortfolio Create s SonarQube Portfolio via WebAPI
func createPortfolio(portfolio string) error { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	apiUrl := "/api/views/create"
	form := url.Values{}
	form.Add("name", portfolio) // Name for the new portfolio
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiUrl, form, "application/x-www-form-urlencoded")
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	var result createResp
	body, err := io.ReadAll(resp.Body)                   // response body is []byte
	if err = json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		return err
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
		return fmt.Errorf("createPortfolio Http Response: %v", resp.StatusCode)
	}
	return nil
}

// addToPortfolio Adds a project to a portfolio via WebAPI
func addToPortfolio(portfolio string, project string) error { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	apiUrl := "/api/views/add_project"
	form := url.Values{}
	form.Add("key", portfolio)   // Key of the portfolio
	form.Add("project", project) // Key of the project
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiUrl, form, "application/x-www-form-urlencoded")
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	var result createResp                                // To be honestly care about the response ?
	body, err := io.ReadAll(resp.Body)                   // response body is []byte
	if err = json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		return err
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
		return fmt.Errorf("addToPortfolio Http Response: %v", resp.StatusCode)
	}

	return nil
}

// setDefaultBranch renames the current default branch of the sonarqube project to the one that is used in GitHub
func setDefaultBranch(repositoryPayload github.RepositoryPayload) error {
	apiUrl := "/api/project_branches/rename"
	form := url.Values{}
	form.Add("name", repositoryPayload.Repository.DefaultBranch)
	form.Add("project", repositoryPayload.Repository.Name)
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiUrl, form, "application/x-www-form-urlencoded")
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
		return fmt.Errorf("setDefaultBranch Http Response: %v", resp.StatusCode)
	}
	return nil
}

// setGitHubBinding links the SonarQube project to the GitHub repository
func setGitHubBinding(repositoryPayload github.RepositoryPayload) error {
	apiUrl := "/api/alm_settings/set_github_binding"
	form := url.Values{}
	form.Add("almSetting", "GitHub")
	form.Add("project", repositoryPayload.Repository.Name)
	form.Add("monorepo", "no")
	form.Add("repository", repositoryPayload.Repository.FullName)
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiUrl, form, "application/x-www-form-urlencoded")
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
		return fmt.Errorf("setGitHubBinding Http Response: %v", resp.StatusCode)
	}

	return nil
}

// OnboardSonarQube bootstraps the SonarQube Plugin
func OnboardSonarQube(repositoryPayload github.RepositoryPayload) { //TODO handle the error from here probably
	search, err := SearchSonarQube(ProjectQualifier, repositoryPayload.Repository.Name)
	if err != nil {
		log.Error(err)
		return
	}
	if search {
		log.Warning("Project " + repositoryPayload.Repository.Name + " Already exists")
		return
	}
	err = createProject(repositoryPayload)
	if err != nil {
		log.Error(err)
		return
	}
	err = setGitHubBinding(repositoryPayload)
	if err != nil {
		log.Error(err)
		return
	}
	err = setDefaultBranch(repositoryPayload)
	if err != nil {
		log.Error(err)
		return
	}
	search, err = SearchSonarQube(PortfolioQualifier, "devops")
	if err != nil {
		log.Error(err)
		return
	}
	if !search {
		createPortfolio("devops")
	}
	addToPortfolio("devops", repositoryPayload.Repository.Name)
	//	handlerGithub.CreateSonarQubeFile(repositoryPayload)
}

//TODO add function that adds the sonar-projects.properties file back to the repo we just onboarded. I think this belongs in the github library
//https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/ -> Configuring your project
