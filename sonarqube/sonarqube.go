package sonarqube

import (
	"bytes"
	"encoding/json"
	"fmt"
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
func DoesProjectExist(repoName string) bool {
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod its just for good development
	mytoken := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	reqstr := SonarUrl + "/api/components/search?qualifiers=TRK&q=" + repoName
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

		if component.Name == repoName || component.Key == repoName || component.Project == repoName { // Catches also some edge cases with weird formatting of names. refactor maybe ?
			return false
		}
	}
	return true
}

func createProject(reponame string) {
	err := godotenv.Load("sonar.env") // This env file needs to be in root. we will remove this during prod its just for good development
	mytoken := os.Getenv("sonartoken")
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	postBody, _ := json.Marshal(map[string]string{ // this could be inplace i guess
		"name":    reponame,
		"project": reponame,
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
