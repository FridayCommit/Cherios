package main

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
	path                 = "/github"
	repoAsCodeOrg        = "FridayCommit"
	repoAsCodeRepository = "as-code"
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

func handleCreateRepositoryEvent(repositoryPayload github.RepositoryPayload) {
	client := initGitHubClient()
	repository := repositoryPayload.Repository
	repositoryJSON, err := json.MarshalIndent(repository, "", "    ")
	if err != nil {
		log.Error("Unable to read repository as JSON")
	}
	message := fmt.Sprintf("Added repo: %s", repository.Name)
	// TODO
	// Add something that overwrites the file or checks if the file exists and then deletes it so we can replace it?
	// If the file exists, call the modify function instead. Update file
	//https://pkg.go.dev/github.com/google/go-github/v48/github@v48.1.0#RepositoryContentFileOptions
	opts := githubApi.RepositoryContentFileOptions{
		Message:   &message,
		Content:   repositoryJSON,
		SHA:       nil,
		Branch:    nil,
		Author:    nil,
		Committer: nil,
	}
	_, _, err = client.Repositories.CreateFile(context.TODO(), repoAsCodeOrg, repoAsCodeRepository, fmt.Sprintf("github/%s.json", repository.Name), &opts)
	if err != nil {
		// TODO: Proper error handling

		return
	}

}

func main() {

	hook, _ := github.New(github.Options.Secret("MyGitHubSuperSecretSecrect...?"))

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.RepositoryEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				log.Warning("Event not supported wallah")
				// ok event wasn;t one of the ones asked to be parsed
			} else {
				log.Warning("Payload not recognized by GitHub wallah")
			}
		}
		switch payload.(type) {

		case github.RepositoryPayload:
			repository := payload.(github.RepositoryPayload)
			switch repository.Action {

			case "created":
				handleCreateRepositoryEvent(repository)
			default:
				log.Warning("Action " + repository.Action + " is not supported")
			}

			//		log.Info(fmt.Printf("%+v", repository))

		}
	})
	log.Info("Listening on port 3000")
	http.ListenAndServe(":3000", nil)

}
