package sonarqube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
)

const (
	SonarUrl = "https://sonarqube.snowdev.io"
)

// TODO
// We need to get or pass the default branch from the github call
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

// Checks if the project already exists in sonarqube
func DoesProjectExist(repositoryPayload github.RepositoryPayload) bool {
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod its just for good development
	mytoken := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	reqstr := SonarUrl + "/api/components/search?qualifiers=TRK&q=" + repositoryPayload.Repository.Name
	req, err := http.NewRequest(http.MethodGet, reqstr, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.SetBasicAuth(mytoken, "") // Leave the password empty, sonarQube is stupid like that
	resp, err := client.Do(req)
	if err != nil {
		//todo
	}
	defer resp.Body.Close()
	var result projectResp
	body, err := io.ReadAll(resp.Body)                    // response body is []byte
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	// The sonarqube API searches partially. So if someone would name their repo devops-app then devops-applications would also show up.
	for _, component := range result.Components {

		if component.Name == repositoryPayload.Repository.Name || component.Key == repositoryPayload.Repository.Name || component.Project == repositoryPayload.Repository.Name { // Catches also some edge cases with weird formatting of names. refactor maybe ?
			return false
		}
	}
	return true
}

func createProject(repositoryPayload github.RepositoryPayload) { // Maybe we should send the repo object instead because other functions might need to know branch etc.
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod its just for good development
	mytoken := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	postBody, _ := json.Marshal(map[string]string{ // this could be inplace i guess
		"name":    repositoryPayload.Repository.Name,
		"project": repositoryPayload.Repository.Name,
	})
	//	responseBody := bytes.NewBuffer(postBody)
	reqstr := SonarUrl + "/api/projects/create" // add these to some kind of types library i guess ?
	req, err := http.NewRequest(http.MethodPost, reqstr, bytes.NewBuffer(postBody))
	req.SetBasicAuth(mytoken, "") // Leave the password empty, sonarQube is stupid like that
	resp, err := client.Do(req)
	if err != nil {
		//todo
	}
	defer resp.Body.Close()
	var result createResp
	body, err := io.ReadAll(resp.Body)                    // response body is []byte
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	fmt.Println(result)
}
func setDefaultBranch(repositoryPayload github.RepositoryPayload) {
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod its just for good development
	mytoken := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	postBody, _ := json.Marshal(map[string]string{ // this could be inplace i guess
		"name":    repositoryPayload.Repository.DefaultBranch,
		"project": repositoryPayload.Repository.Name,
	})
	//	responseBody := bytes.NewBuffer(postBody)
	reqstr := SonarUrl + "/api/project_branches/rename" // add these to some kind of types library i guess ?
	req, err := http.NewRequest(http.MethodPost, reqstr, bytes.NewBuffer(postBody))
	req.SetBasicAuth(mytoken, "") // Leave the password empty, sonarQube is stupid like that
	resp, err := client.Do(req)
	if err != nil {
		//todo
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		//todo
	}
}
func setGitHubBinding(repositoryPayload github.RepositoryPayload) {
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod its just for good development
	mytoken := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	postBody, _ := json.Marshal(map[string]string{ // this could be inplace i guess
		"almSetting": "GitHub",
		"project":    repositoryPayload.Repository.Name,
		"monorepo":   "no",
		"repository": repositoryPayload.Repository.FullName, //Needs to be Org/RepoName
	})
	//	responseBody := bytes.NewBuffer(postBody)
	reqstr := SonarUrl + "/api/alm_settings/set_github_binding" // add these to some kind of types library i guess ?
	req, err := http.NewRequest(http.MethodPost, reqstr, bytes.NewBuffer(postBody))
	req.SetBasicAuth(mytoken, "") // Leave the password empty, sonarQube is stupid like that
	resp, err := client.Do(req)
	if err != nil {
		//todo
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		//todo
	}
}
