package main

import (
	"fmt"
	"os"
	"strings"

	jira "github.com/andygrunwald/go-jira"
)

func init() {
	// For now pull the required auth information from the environment.
	//
	// In order for us to correctly interact with Jira, we'll presume that
	// we need authentication information (i.e. it's not a public JIra
	// deployment). This means that we need:
	//
	//  - JIRA_USER
	//  - JIRA_API_TOKEN
	//  - JIRA_BASE_URL

	jiraUser := os.Getenv("JIRA_USER")
	if len(jiraUser) == 0 {
		printHelpAndExit("JIRA_USER")
	}

	apiToken := os.Getenv("JIRA_API_TOKEN")
	if len(apiToken) == 0 {
		printHelpAndExit("JIRA_API_TOKEN")
	}

	jiraBaseURL := os.Getenv("JIRA_BASE_URL")
	if len(jiraBaseURL) == 0 {
		printHelpAndExit("JIRA_BASE_URL")
	}

	// Note that this is a misnomer - the `Password` can be either
	// the user's password, *not recommended*, or it can be an API
	// token for the user that can be created by following these
	// instructions:
	// https://confluence.atlassian.com/cloud/api-tokens-938839638.html#APItokens-CreateanAPItoken
	tp := jira.BasicAuthTransport{
		Username: jiraUser,
		Password: apiToken,
	}

	client, err := jira.NewClient(tp.Client(), jiraBaseURL)
	if err != nil {
		panic(err)
	}

	jiraClient = client

	// One can also, optionally, pass in a list of allowed CORS origins
	corsOrigins := os.Getenv("ALLOWED_CORS_ORIGINS")
	if len(corsOrigins) > 0 {
		allowedCORSOrigins = strings.Split(corsOrigins, ",")
	}

	// We can also modify the port and interface that we listen on, by
	// default it's :9090.
	httpAddrEnv := os.Getenv("HTTP_ADDR")
	if len(httpAddrEnv) > 0 {
		httpAddr = httpAddrEnv
	}
}

// printHelpAndExit informs the user that they need to provide the given
// environment variable and then exits with setting the exit status to 1.
func printHelpAndExit(missingEnvvar string) {
	fmt.Printf("must set %q\n", missingEnvvar)
	os.Exit(1)
}
