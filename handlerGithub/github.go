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
	repoAsCode           = repoAsCodeOrg + "/" + repoAsCodeRepository
	appID                = 263646 // https://github.com/apps/cheriosapp
)

type RenameChangesPayload struct {
	Changes struct {
		Repository struct {
			Name struct {
				From string `json:"from"`
			} `json:"name"`
		} `json:"repository"`
	} `json:"changes"`
}

type RepoSchema struct {
	Name   string   `json:"Name"`
	People []People `json:"People"`
}
type People struct {
	Name string `json:"Name"`
	Role string `json:"Role"`
}

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

func convertToGithubRepositorySchema(repositoryPayload github.RepositoryPayload) *RepoSchema {
	githubRepository := RepoSchema{
		Name: repositoryPayload.Repository.Name,
		People: []People{
			{
				Name: "Felix",
				Role: "Admin",
			},
			{
				Name: "Viktor",
				Role: "Admin",
			},
		},
	}

	return &githubRepository
}

func createFile(client *githubApi.Client, opts githubApi.RepositoryContentFileOptions, filePath string) {
	repositoryContentResponse, _, err := client.Repositories.CreateFile(context.TODO(), repoAsCodeOrg, repoAsCodeRepository, filePath, &opts)
	if err != nil {
		// TODO: Proper error handling
		return
	}
	log.Info(fmt.Sprintf("File %s/%s created in commit %s", repoAsCode, filePath, *repositoryContentResponse.Commit.SHA))
}

func updateFile(client *githubApi.Client, opts githubApi.RepositoryContentFileOptions, filePath string) {
	repositoryContentResponse, _, err := client.Repositories.UpdateFile(context.TODO(), repoAsCodeOrg, repoAsCodeRepository, filePath, &opts)
	if err != nil {
		// TODO: Proper error handling
		return
	}
	log.Info(fmt.Sprintf("File %s/%s updated in commit %s", repoAsCode, filePath, *repositoryContentResponse.Commit.SHA))
}

//TODO maybe public
func deleteFile(client *githubApi.Client, opts githubApi.RepositoryContentFileOptions, filePath string) {
	repositoryContentResponse, _, err := client.Repositories.DeleteFile(context.TODO(), repoAsCodeOrg, repoAsCodeRepository, filePath, &opts)
	if err != nil {
		// TODO: Proper error handling
		return
	}
	log.Info(fmt.Sprintf("File %s/%s deleted in commit %s", repoAsCode, filePath, *repositoryContentResponse.Commit.SHA))
}

func getFile(path string, client *githubApi.Client) (*githubApi.RepositoryContent, bool) {
	opts := githubApi.RepositoryContentGetOptions{}
	fileContent, _, resp, err := client.Repositories.GetContents(context.TODO(), repoAsCodeOrg, repoAsCodeRepository, path, &opts)
	if err != nil {
		// TODO: Proper error handling
		return nil, false
	}
	return fileContent, resp.StatusCode == http.StatusOK
}

func HandleRepositoryEvent(repositoryPayload github.RepositoryPayload, renameChangePayload RenameChangesPayload) {
	client := initGitHubClient()

	repositorySchema := convertToGithubRepositorySchema(repositoryPayload)
	repositoryJSON, err := json.MarshalIndent(repositorySchema, "", "    ")
	if err != nil {
		log.Error("Unable to read repository as JSON")
	}

	filePath := fmt.Sprintf("github/%s.json", repositoryPayload.Repository.Name)
	message := fmt.Sprintf("Update GitHub repo %s", repositoryPayload.Repository.Name)
	fileContent, exists := getFile(filePath, client)
	var sha *string = nil
	if fileContent != nil {
		sha = fileContent.SHA
	}
	opts := githubApi.RepositoryContentFileOptions{
		Message:   &message,
		Content:   repositoryJSON,
		SHA:       sha,
		Branch:    nil,
		Author:    nil,
		Committer: nil,
	}
	// Event doc here: https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads#repository
	switch repositoryPayload.Action {
	case "created":
		if exists { //If the file already exist for some reason, we update it instead of creating. Catch
			updateFile(client, opts, filePath)
		} else {
			createFile(client, opts, filePath)
		}
	case "edited":
		// TODO: Handle changes of topics, default branch, description, or homepage of a repository was changed ()
		updateFile(client, opts, filePath)
	case "deleted":
		deleteFile(client, opts, filePath)
	case "renamed":
		/**
		// TODO When Changing the name of the Repo
		oldRepoName := renameChangePayload.Changes.Repository.Name
		fileContent, exists := getFile(fmt.Sprintf("github/%s", oldRepoName), client)
		if !exists {
			log.Errorf("File %s in %s not found", filePath, repoAsCode)
		}

		client.Git.CreateTree(context.TODO(), repoAsCodeOrg, repoAsCodeRepository)
		*/
		// Actions should be 1. Get the file 2. Populate our schema with the source schema 3. Add changes 4. Push
		updateFile(client, opts, filePath)
	default:
		log.Warning("Action " + repositoryPayload.Action + " is not supported")
	}

}

/*
Obmaa
obama2


*/

// Creates a hook to the source of truth repo so that we can see changes to files. Can be ran on init
func CreateSourceHook() {
	//	var interfaceVal interface{}
	//	json.Unmarshal(j, &interfaceVal)
	test := map[string]interface{}{
		"url":          "http://84.216.123.207:3000/github",
		"content_type": "json",
		"insecure_ssl": 0,
		"secret":       "MyGitHubSuperSecretSecrect...?",
	}
	hookName := "SourceHook"
	activeBool := true
	opts := githubApi.Hook{
		CreatedAt:    nil,
		UpdatedAt:    nil,
		URL:          nil,
		ID:           nil,
		Type:         nil,
		Name:         &hookName,
		TestURL:      nil,
		PingURL:      nil,
		LastResponse: nil,
		Events:       []string{"push"},
		Config:       test,
		Active:       &activeBool,
	}
	client := initGitHubClient()
	hook, resp, err := client.Repositories.CreateHook(context.TODO(), repoAsCodeOrg, repoAsCodeRepository, &opts)
	if err != nil {
		//TODO fix error
		return
	}
	log.Info(fmt.Sprintf("Hook response %s"), resp.Status)
	log.Info(fmt.Sprintf("Hook %s created for %s"), hook.Name, repoAsCode)

}
