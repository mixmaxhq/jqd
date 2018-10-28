package main

import (
	jira "github.com/andygrunwald/go-jira"
	"github.com/sirupsen/logrus"
)

// Configured in init.go (on startup).
var jiraClient *jira.Client

func main() {
	logrus.Info("starting server on ", httpAddr)
	go startRouter()

	shutdown := setupShutdownHooks()

	<-shutdown
}
