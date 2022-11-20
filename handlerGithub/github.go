package handlerGithub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-playground/webhooks/v6/github"
	githubApi "github.com/google/go-github/v48/github"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var appKey string

const (
	Path                 = "/github"
	repoAsCodeOrg        = "FridayCommit"
	repoAsCodeRepository = "as-code"
	repoAsCode = repoAsCodeOrg + "/" + repoAsCodeRepository
	appID                = 263646 // https://github.com/apps/cheriosapp
)

func initGitHubClient() *githubApi.Client {
	//	if len(appKey) < 1 {
	//		log.Fatalln("Missing App Key")
	//	}
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appID, 31393521, "cheriosapp.2022-11-19.private-key.pem")
	//	itr, err := ghinstallation.New(http.DefaultTransport, 250575, 30374345, []byte(appKey))
	if err != nil {
		log.Fatalln(err)
	}
	client := githubApi.NewClient(&http.Client{Transport: itr})
	return client
}


type githubRepositorySchema struct {
    Name string
    People []struct {
		Name string
		Role string
	}
}

func convertToGithubRepositorySchema(repositoryPayload github.RepositoryPayload) *githubRepositorySchema {
	githubRepository := githubRepositorySchema{
		Name: repositoryPayload.Repository.Name,
		People: []struct{Name string; Role string}{
			{Name: "Felix", Role:  "Admin"},
		},
	}

	return &githubRepository
}

func HandleCreateRepositoryEvent(repositoryPayload github.RepositoryPayload) {
	client := initGitHubClient()

	repositorySchema := convertToGithubRepositorySchema(repositoryPayload)
	repositoryJSON, err := json.MarshalIndent(repositorySchema, "", "    ")
	if err != nil {
		log.Error("Unable to read repository as JSON")
	}

	// TODO
	// Add something that overwrites the file or checks if the file exists and then deletes it so we can replace it?
	// If the file exists, call the modify function instead. Update file
	//https://pkg.go.dev/github.com/google/go-github/v48/github@v48.1.0#RepositoryContentFileOptions
	message := fmt.Sprintf("Update GitHub repo %s", repositoryPayload.Repository.Name)
	filePath := fmt.Sprintf("github/%s.json", repositoryPayload.Repository.Name)
	opts := githubApi.RepositoryContentFileOptions{
		Message:   &message,
		Content:   repositoryJSON,
		SHA:       nil,
		Branch:    nil,
		Author:    nil,
		Committer: nil,
	}
	repositoryContentResponse, _, err := client.Repositories.CreateFile(context.TODO(), repoAsCodeOrg, repoAsCodeRepository, filePath, &opts)
	if err != nil {
		// TODO: Proper error handling
		return
	}
	log.Info(fmt.Sprintf("File %s/%s updated in commit %s", repoAsCode, filePath, *repositoryContentResponse.Commit.SHA))
}
