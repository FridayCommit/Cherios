package main

//TODO
// Implement a Schema with the most important fields for repo
// Make Cherios check and create webhook to the as-code repo if needed

// TODO Environments
// Make a function that checks the required variables ? maybe docker can do that :)
import (
	"bytes"
	"encoding/json"
	"fridaycommit/cherios/handlerGithub"
	"github.com/go-playground/webhooks/v6/github"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

const ( // Move some of these to Inputs instead they shouldn't be Constants. except maybe the webhook path ?
	path = "/github"
)

func init() {
	// Bootstrap
	err := godotenv.Load("settings.env") // Includes our settings such as enable sonarqube and config for RepoAsCode
	if err != nil {
		log.Fatalln(err)
	}
	err = godotenv.Load("secrets.env") // This env file needs to be in root. we will remove this during prod it's just for good development
	if err != nil {
		log.Fatalln(err)
	}
	err = handlerGithub.CreateSourceHook()
	if err != nil {
		log.Fatalln(err)
	}
}

func ParseRenameChangeHook(r *http.Request) (handlerGithub.RenameChangesPayload, error) {
	payload, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Error("Failed to read request body")
	}
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	var pl handlerGithub.RenameChangesPayload
	err = json.Unmarshal([]byte(payload), &pl)
	if err != nil {
		log.Error("Failed to parse body to json")
	}
	return pl, err
}

func main() {

	hook, _ := github.New(github.Options.Secret("MyGitHubSuperSecretSecrect...?"))

	http.HandleFunc(handlerGithub.Path, func(w http.ResponseWriter, r *http.Request) {
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(r.Body)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		payload, err := hook.Parse(r, github.RepositoryEvent) // This function closes the body thats why it dosent work
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
			var renameChangePayload handlerGithub.RenameChangesPayload

			repository := payload.(github.RepositoryPayload)
			if repository.Action == "renamed" { // TODO this function is broken we need to fix it
				r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
				renameChangePayload, err = ParseRenameChangeHook(r)
			}
			handlerGithub.HandleRepositoryEvent(repository, renameChangePayload)
		default:
			log.Warning("Something went wrong")
		}

	})
	log.Info("Listening on port 3000")
	http.ListenAndServe(":3000", nil)

}
