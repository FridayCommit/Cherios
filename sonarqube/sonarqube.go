package sonarqube

import (
	"encoding/json"
	"fmt"
	"fridaycommit/cherios/handlerGithub"
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
	apiSearch             = "/api/components/search"
	apiProjectCreate      = "/api/projects/create"
	apiSetGitHubBinding   = "/api/alm_settings/set_github_binding"
	apiRenameBranch       = "/api/project_branches/rename"
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

func sonarqubeCall(method string, requeststr string, form url.Values, contentType string) (*http.Response, error) {
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod its just for good development
	if err != nil {
		//TODO
	}
	mytoken := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	req, err := http.NewRequest(method, requeststr, strings.NewReader(form.Encode()))
	req.SetBasicAuth(mytoken, "")
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	resp, err := client.Do(req)
	if err != nil {
	}
	return resp, nil
}

// SearchSonarQube Checks if the project already exists in sonarqube
func SearchSonarQube(qualifier string, search string) bool {
	urlA, err := url.Parse(apiSearch) // TODO this is a little ugly but it works super well. maybe works with forms
	if err != nil {
		log.Fatal(err)
	}
	values := urlA.Query()
	values.Add("qualifiers", qualifier) // enum maybe? maybe not needed
	values.Add("q", search)
	urlA.RawQuery = values.Encode()
	resp, err := sonarqubeCall(http.MethodGet, SonarUrl+urlA.String(), nil, "application/json; charset=UTF-8")
	defer resp.Body.Close()
	if err != nil {
		// TODO
	}
	var result searchResp
	body, err := io.ReadAll(resp.Body)
	if err = json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	// The sonarqube API searches partially. So if someone would name their repo devops-app then devops-applications would also show u<p.
	for _, component := range result.Components { // TODO we could break this out to its own function i guess ?
		fields := reflect.ValueOf(component)
		for i := 0; i < fields.NumField(); i++ {
			if fields.Field(i).String() == search {
				return true
			}
		}
	}
	return false
}

// createProject Creates a SonarQube Project via WebAPI
func createProject(repositoryPayload github.RepositoryPayload) { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	form := url.Values{}
	form.Add("name", repositoryPayload.Repository.Name)
	form.Add("project", repositoryPayload.Repository.Name)
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiProjectCreate, form, "application/x-www-form-urlencoded")
	if err != nil {
		// TODO
	}
	var result createResp
	body, err := io.ReadAll(resp.Body)                    // response body is []byte
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
	}
	defer resp.Body.Close()
}

// createPortfolio Create s SonarQube Portfolio via WebAPI
func createPortfolio(portfolio string) { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	apiUrl := "api/views/create"
	form := url.Values{}
	form.Add("name", portfolio) // Name for the new portfolio
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiUrl, form, "application/x-www-form-urlencoded")
	if err != nil {
		// TODO
	}
	var result createResp
	body, err := io.ReadAll(resp.Body)                    // response body is []byte
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
	}
	defer resp.Body.Close()
}

// addToPortfolio Adds a project to a portfolio via WebAPI
func addToPortfolio(portfolio string, project string) { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	apiUrl := "api/views/add_project"
	form := url.Values{}
	form.Add("key", portfolio)   // Key of the portfolio
	form.Add("project", project) // Key of the project
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiUrl, form, "application/x-www-form-urlencoded")
	if err != nil {
		// TODO
	}
	var result createResp                                 // To be honestly care about the response ?
	body, err := io.ReadAll(resp.Body)                    // response body is []byte
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
	}
	defer resp.Body.Close()
}

// setDefaultBranch renames the current default branch of the sonarqube project to the one that is used in GitHub
func setDefaultBranch(repositoryPayload github.RepositoryPayload) {
	form := url.Values{}
	form.Add("name", repositoryPayload.Repository.DefaultBranch)
	form.Add("project", repositoryPayload.Repository.Name)
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiRenameBranch, form, "application/x-www-form-urlencoded")
	if err != nil {
		//TODO
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
	}
	defer resp.Body.Close()
}

// setGitHubBinding links the SonarQube project to the GitHub repository
func setGitHubBinding(repositoryPayload github.RepositoryPayload) {
	form := url.Values{}
	form.Add("almSetting", "GitHub")
	form.Add("project", repositoryPayload.Repository.Name)
	form.Add("monorepo", "no")
	form.Add("repository", repositoryPayload.Repository.FullName)
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiSetGitHubBinding, form, "application/x-www-form-urlencoded")
	if err != nil {
		// TODO
	}
	if resp.StatusCode > 299 {
		log.Warning(resp.StatusCode)
	}
	defer resp.Body.Close()
}

// OnboardSonarQube bootstraps the SonarQube Plugin
func OnboardSonarQube(repositoryPayload github.RepositoryPayload) {
	if SearchSonarQube(ProjectQualifier, repositoryPayload.Repository.Name) {
		log.Warning("Project " + repositoryPayload.Repository.Name + " Already exists")
		return
	}
	createProject(repositoryPayload)
	setGitHubBinding(repositoryPayload)
	setDefaultBranch(repositoryPayload)
	handlerGithub.CreateSonarQubeFile(repositoryPayload)
}

//TODO add function that adds the sonar-projects.properties file back to the repo we just onboarded. I think this belongs in the github library
//https://docs.sonarqube.org/latest/analysis/scan/sonarscanner/ -> Configuring your project
