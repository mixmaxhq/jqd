package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dimfeld/httptreemux"
	"github.com/gorilla/schema"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

var (
	allowedCORSOrigins = []string{"*"}
	httpAddr           = ":9090"
)

// startRouter starts the web server with our API route.
func startRouter() {
	mux := httptreemux.NewContextMux()

	mux.GET("/api/search", searchHandler)

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedCORSOrigins,
		AllowCredentials: true,
	})

	http.ListenAndServe(httpAddr, c.Handler(mux))
}

// searchHandler handles requests to the /api/search endpoint.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Parse our search parameters.
	var params SearchParams
	dec := schema.NewDecoder()
	dec.RegisterConverter(time.Now(), JSONTimeConverter)
	if err := dec.Decode(&params, r.URL.Query()); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logrus.Info("failed to decode query params, err: ", err)
		return
	}

	// Construct our Jira JQL query.
	query := queryForParams(params)

	// Retrieve the search results from Jira
	results, err := searchIssues(query, params.GroupBy)
	if err != nil {
		logrus.Info("failed to search issues, err: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var toReturn interface{} = results

	// If the user only wanted aggregate stats, aggregate those now.
	if _, ok := r.URL.Query()["aggregate"]; ok {
		toReturn = aggregateResults(results)
	}

	encoder := json.NewEncoder(w)

	// If the user wanted a pretty JSON response, make sure that we set a
	// JSON indent.
	if _, ok := r.URL.Query()["pretty"]; ok {
		encoder.SetIndent("", "  ")
	}

	w.Header().Add("Content-Type", "application/json")

	if err := encoder.Encode(toReturn); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// NamedCount is data pair of a label's `Name` and the `Count` of issues
// tagged with said label.
type NamedCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// AggregateResults is the list of all `NamedCount` pairs that appeared
// in the search results.
type AggregateResults struct {
	Data []NamedCount `json:"data"`
}

// aggregateResults sums the individual timeseries counts for the given
// results and returns them.
func aggregateResults(r *Results) *AggregateResults {
	agg := AggregateResults{
		Data: make([]NamedCount, len(r.Data)),
	}

	for i, r := range r.Data {
		count := 0

		for _, p := range r.Values {
			count += p
		}

		agg.Data[i] = NamedCount{
			Name:  r.Name,
			Count: count,
		}
	}

	return &agg
}
