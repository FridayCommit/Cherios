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
	"strings"
)

const (
	apiSearch           = "/api/components/search?qualifiers=TRK&q="
	apiCreate           = "/api/projects/create"
	apiSetGitHubBinding = "/api/alm_settings/set_github_binding"
	apiRenameBranch     = "/api/project_branches/rename"
)

var (
	SonarUrl = "https://sonarqube.snowdev.io"
)

// TODO
// We need to get or pass the default branch from the github call
// TODO ADD ERROR HANDLING FOR EVERYTHING
type projectResp struct {
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
		//todo
	}
	return resp, nil
}

// DoesProjectExist Checks if the project already exists in sonarqube
func DoesProjectExist(repositoryPayload github.RepositoryPayload) bool {
	resp, err := sonarqubeCall(http.MethodGet, SonarUrl+apiSearch+repositoryPayload.Repository.Name, nil, "application/json; charset=UTF-8")
	if err != nil {
		// TODO
	}
	var result projectResp
	body, err := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Println("Can not unmarshal JSON")
	}
	// The sonarqube API searches partially. So if someone would name their repo devops-app then devops-applications would also show up.
	for _, component := range result.Components {
		if component.Name == repositoryPayload.Repository.Name || component.Key == repositoryPayload.Repository.Name || component.Project == repositoryPayload.Repository.Name { // Catches also some edge cases with weird formatting of names. refactor maybe ?
			return true
		}
	}
	defer resp.Body.Close()
	return false
}
func createProject(repositoryPayload github.RepositoryPayload) { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	form := url.Values{}
	form.Add("name", repositoryPayload.Repository.Name)
	form.Add("project", repositoryPayload.Repository.Name)
	resp, err := sonarqubeCall(http.MethodPost, SonarUrl+apiCreate, form, "application/x-www-form-urlencoded")
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
func setDefaultBranch(repositoryPayload github.RepositoryPayload) {
	//	postBody, _ := json.Marshal(map[string]string{ // this could be inplace i guess
	//		"name":    repositoryPayload.Repository.DefaultBranch,
	//		"project": repositoryPayload.Repository.Name,
	//	})
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
func OnboardSonarQube(repositoryPayload github.RepositoryPayload) {
	if DoesProjectExist(repositoryPayload) {
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
