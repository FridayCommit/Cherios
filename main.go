package main

//TODO
// Implement a Schema with the most important fields for repo
// Make Cherios check and create webhook to the as-code repo if needed
import (
	"fridaycommit/cherios/handlerGithub"

	"github.com/go-playground/webhooks/v6/github"
	log "github.com/sirupsen/logrus"
	"net/http"
)

var appKey string

const (
	path                 = "/github"
	repoAsCodeOrg        = "FridayCommit"
	repoAsCodeRepository = "as-code"
	repoAsCode = repoAsCodeOrg + "/" + repoAsCodeRepository
	appID                = 263646 // https://github.com/apps/cheriosapp
)

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
			repository := payload.(github.RepositoryPayload)
			switch repository.Action {

			case "created":
				handlerGithub.HandleCreateRepositoryEvent(repository)
			default:
				log.Warning("Action " + repository.Action + " is not supported")
			}

			//		log.Info(fmt.Printf("%+v", repository))

		}
	})
	log.Info("Listening on port 3000")
	http.ListenAndServe(":3000", nil)

}
