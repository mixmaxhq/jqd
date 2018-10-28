# jqd
`jqd` is a Jira query daemon that exposes the ability to search for Jira
issues in a rich manner and to visualize those over different time periods.


## Get it
### Docker
If you want to run `jqd` via docker, you can:

```sh
docker pull mixmaxhq/jqd
```


### Source
If you want to get it and run it from source, you can:

```
go get github.com/mixmaxhq/jqd
dep ensure
go run main.go
```

## Running `jqd`

### Required values

To run `jqd`, you need to provide it with some information. In order to
communicate with Jira, `jqd` needs to know:

 - The username of a Jira account to connect with
 - Either an API token or the password for that account, please only use
   API tokens - NEVER use passwords.
 - The base URL for the Jira deployment to communicate with.
 
If running `jqd` from source, this will need to be provided via the environment
for now as:

 - $JIRA_USER
 - $JIRA_API_TOKEN (this _can_ be a password, but please don't)
 - $JIRA_BASE_URL
 
If running via docker, you can do so as:

```sh
docker run -d -p 9090:9090            \
    -e JIRA_USER=$JIRA_USER           \
    -e JIRA_API_TOKEN=$JIRA_API_TOKEN \
    -e JIRA_BASE_URL=$JIRA_BASE_URL   \
    mixmaxhq/jqd:latest
```

## Searching

Currently, `jqd` exposes a single API endpoint at `/api/search`. This API
endpoint only supports GET requests and responds with a JSON resource. It
supports search parameters via query parameters. The supported values are:

| Value | Details |
| ----- | ------- |
| labels | A single label or multiple labels to constrain the search by. |
| components | A single component or multiple components to constrain the search by. |
| project | A single project identifier to constrain the search by. |
| status | The status of tickets to constrain the search by. |
| statusCategory | The status category to constrain the search by. |
| issueType | The issue type to constrain the search by. |
| createdAfter | The creation start date to constrain the search by. |
| createdBefore | The creation end date to constrain the search by. |
| rawQuery | A raw JQL query to run. |

There are optional parameters to mutate the response with.
| Value | Details |
| ----- | ------- |
| pretty | Pretty prints the JSON response. |
| aggregate | Aggregates the counts per label. |


## Developing

If you'd like to contribute to `jqd`, there's a utility function in the
`Makefile` to make this easier, once done modifying the code you can run

```sh
make test
```

Which will rebuild the binary, then a new docker image and will launch it,
pulling in the needed envvars.
