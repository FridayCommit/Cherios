package sonarqube

import (
	"net/http"
)

const (
	SonarUrl = "https://sonarqube.com"
	mytoken  = "Mytokebhere?"
)

// TODO
// We need to get or pass the default branch from the github call
//

// Checks if the project already exists in sonarqube
func doesProjectExist(repoName string) {
	var bearer = "Bearer " + mytoken
	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	req, err := http.NewRequest(http.MethodGet, SonarUrl+"/api/components/search?qualifiers=TRK&q="+repoName, nil)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Authorization", bearer)
	resp, err := client.Do(req)
	if err != nil {
		//todo
	}
	defer resp.Body.Close()

}
