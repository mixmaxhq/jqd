package main

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
)

// listComponents lists the components for the given project.
func listComponents(project string) ([]string, error) {

	// go-jira doesn't have this yet, so make the request ourselves.
	req, _ := jiraClient.NewRequest("GET", fmt.Sprintf(
		"rest/api/3/project/%s/components",
		project,
	), nil)

	var components []jira.Component
	_, err := jiraClient.Do(req, &components)
	if err != nil {
		return nil, err
	}

	// Slim down the components to just their names.
	names := make([]string, len(components))
	for i, comp := range components {
		names[i] = comp.Name
	}

	return names, nil
}
