package main

//TODO
// Implement a Schema with the most important fields for repo
// Make Cherios check and create webhook to the as-code repo if needed
import (
	"bytes"
	"encoding/json"
	"fridaycommit/cherios/handlerGithub"
	"fridaycommit/cherios/sonarqube"
	"github.com/go-playground/webhooks/v6/github"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

const ( // Move some of these to Inputs instead they shouldnt be Constants. except maybe the webhook path ?
	path                 = "/github"
	repoAsCodeOrg        = "FridayCommit"
	repoAsCodeRepository = "as-code"
	repoAsCode           = repoAsCodeOrg + "/" + repoAsCodeRepository
	appID                = 263646 // https://github.com/apps/cheriosapp
)

func init() {
	test := github.RepositoryPayload{}
	test.Repository.Name = "devops-jeskai"
	test.Repository.DefaultBranch = "main"
	test.Repository.FullName = "SnowSoftwareGlobal/devops-jeskai"
	sonarqube.SearchSonarQube(sonarqube.ProjectQualifier, test.Repository.Name)
	sonarqube.OnboardSonarQube(test)
	time.Sleep(1000)
	handlerGithub.CreateSourceHook()
}

func ParseRenameChangeHook(r *http.Request) (handlerGithub.RenameChangesPayload, error) {
	payload, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Println("Failed to read request body")
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(payload))
	var pl handlerGithub.RenameChangesPayload
	err = json.Unmarshal([]byte(payload), &pl)
	return pl, err
}

func main() {

	hook, _ := github.New(github.Options.Secret("MyGitHubSuperSecretSecrect...?"))

	http.HandleFunc(handlerGithub.Path, func(w http.ResponseWriter, r *http.Request) {
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
			renameChangePayload, err := ParseRenameChangeHook(r)
			if err != nil {
				log.Warning(err)
			}
			payload, err := hook.Parse(r, github.WorkflowJobEvent, github.PullRequestEvent)
			repository := payload.(github.RepositoryPayload)
			handlerGithub.HandleRepositoryEvent(repository, renameChangePayload)
			sonarqube.OnboardSonarQube(repository)
		default:
			log.Warning("Something went wrong")
		}

	})
	log.Info("Listening on port 3000")
	http.ListenAndServe(":3000", nil)

}
