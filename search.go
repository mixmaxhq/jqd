package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	jira "github.com/andygrunwald/go-jira"
	"github.com/ttacon/jiraquery"
)

// searchIssues searches Jira with the given query.
func searchIssues(query string) (*Results, error) {
	var (
		issueCategories = map[string]map[string]int{}
	)

	// Let's search for those issues! Note that we don't currently have a
	// way to identify how many issues we may process here, so
	// https://giphy.com/gifs/OCu7zWojqFA1W/html5.
	err := jiraClient.Issue.SearchPages(query,
		&jira.SearchOptions{},
		func(i jira.Issue) error {
			// If there aren't any labels, don't do anything at the
			// moment.
			if len(i.Fields.Labels) == 0 {
				return nil
			}

			// Process the issues labels.
			for _, label := range i.Fields.Labels {
				label = strings.TrimSpace(label)

				issues, ok := issueCategories[label]
				if !ok {
					issues = map[string]int{}
					issueCategories[label] = issues
				}
				year, month, day := time.Time(i.Fields.Created).Date()
				timeKey := fmt.Sprintf("%d:%d:%d", year, month, day)

				count, _ := issues[timeKey]
				issues[timeKey] = count + 1
			}
			return nil
		})

	if err != nil {
		return nil, err
	}

	// Clean up the data by stretching out all data points across the
	// maximum time range.
	results := sanitizeData(issueCategories)

	return results, err
}

// queryForParams constructs the desired JQL query from the given search
// parameters.
func queryForParams(params SearchParams) string {
	if len(params.RawQuery) > 0 {
		return params.RawQuery
	}

	builder := jiraquery.AndBuilder()

	if len(params.Project) > 0 {
		builder.Project(params.Project)
	}

	if len(params.IssueType) > 0 {
		builder.IssueType(params.IssueType)
	}

	if len(params.Status) > 0 {
		builder.Eq(jiraquery.Word("status"), jiraquery.Word(params.Status))
	}

	if len(params.StatusCategory) > 0 {
		builder.Eq(
			jiraquery.Word("statusCategory"),
			jiraquery.Word(fmt.Sprintf("%q", params.StatusCategory)))
	}

	if len(params.Labels) > 0 {
		if len(params.Labels) == 1 {
			builder.Eq(jiraquery.Word("labels"), jiraquery.Word(params.Labels[0]))
		} else {
			builder.In(jiraquery.Word("labels"), jiraquery.List(params.Labels...))
		}
	}

	if len(params.Components) > 0 {
		if len(params.Components) == 1 {
			builder.Eq(jiraquery.Word("component"), jiraquery.Word(params.Components[0]))
		} else {
			builder.In(jiraquery.Word("component"), jiraquery.List(params.Components...))
		}
	}

	if params.CreatedAfter != nil {
		builder.GreaterThan(
			jiraquery.Word("created"),
			jiraquery.Word(fmt.Sprintf("%q", params.CreatedAfter.Format("2006-1-2 04:05"))))
	}

	if params.CreatedBefore != nil {
		builder.LessThan(
			jiraquery.Word("created"),
			jiraquery.Word(fmt.Sprintf("%q", params.CreatedBefore.Format("2006-1-2 04:05"))))
	}

	return builder.Value().String()
}

// SearchParams holds top level search parameters while also allowing for more
// complex queries to be passed in via the `rawQuery` field.
type SearchParams struct {
	Labels         []string   `json:"labels"`
	Components     []string   `json:"components"`
	Project        string     `json:"project"`
	Status         string     `json:"status"`
	StatusCategory string     `json:"statusCategory"`
	IssueType      string     `json:"issueType"`
	CreatedBefore  *time.Time `json:"createdBefore"`
	CreatedAfter   *time.Time `json:"createdAfter"`
	RawQuery       string     `json:"rawQuery"`

	// Options
	Aggregate bool `json:"aggregate"`
	Pretty    bool `json:"pretty"`
}

func JSONTimeConverter(raw string) reflect.Value {
	val, err := time.Parse("2006-1-2", raw)
	if err != nil {
		return reflect.Value{}
	}

	return reflect.ValueOf(val)
}

// JSONTime is a helper type for simplifying parsing and embedding Go
// time.Time references.
type JSONTime struct {
	time.Time
}

func (t *JSONTime) MarshalJSON() ([]byte, error) {
	if t == nil {
		return nil, nil
	}

	stamp := fmt.Sprintf("\"%s\"", t.Format("2006-1-2"))
	return []byte(stamp), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	cleanedData := strings.Replace(string(data), "\"", "", -1)
	var err error
	t.Time, err = time.Parse("2006-1-2", cleanedData)
	return err
}

// timeValuePair is a per time based count that we'll use to sort and fill in
// any time series data per label.
type timeValuePair struct {
	time time.Time
	val  int
}

// ByTime is how we'll sort label issue counts per day. Note that we'll sort
// the timeseries points from earliest to most recent.
type ByTime []timeValuePair

func (p ByTime) Len() int           { return len(p) }
func (p ByTime) Less(i, j int) bool { return p[i].time.Before(p[j].time) }
func (p ByTime) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// labelTimeSeriesCollector is a collector entity that we'll use to help us
// fill in timeseries points that may be missing (due to no issues of a given
// type being filed on a given day).
type labelTimeSeriesCollector struct {
	label string
	data  []timeValuePair
}

// ChartSeries is a composite data type containing the name of a Jira label
// and counts of issues with that type.
type ChartSeries struct {
	Name   string `json:"name"`
	Values []int  `json:"values"`
}

// sanitizeData cleans up our analyzed data (better comment coming soon).
func sanitizeData(data map[string]map[string]int) *Results {
	var (
		// We need to generate our series
		dataSeries []ChartSeries

		// Needed for filling in missing data points.
		rawData []labelTimeSeriesCollector

		// Needed for identifying the time range our time series needs
		// to span.
		maxTime time.Time
		minTime time.Time
	)

	const timeFormat = "2006:1:2"

	// "Process them!"
	for label, series := range data {
		ltsc := labelTimeSeriesCollector{label, nil}

		var (
			pairValues = make([]timeValuePair, len(series))
			ii         = 0
		)

		// Process the series data tracking the min and max known times.
		for tPoint, count := range series {
			tim, err := time.Parse(timeFormat, tPoint)
			if err != nil {
				panic(err)
			}

			pairValues[ii] = timeValuePair{tim, int(count)}
			ii++

			if maxTime.IsZero() || maxTime.Before(tim) {
				maxTime = tim
			}

			if minTime.IsZero() || minTime.After(tim) {
				minTime = tim
			}
		}

		ltsc.data = pairValues

		// Store that data for later...
		rawData = append(rawData, ltsc)
	}

	// Synthesize the known time range as a mapping for easy querying.
	var fullTimes = map[string]time.Time{}
	for !minTime.After(maxTime) {
		fullTimes[minTime.Format("2006:1:2")] = minTime

		minTime = minTime.Add(time.Hour * 24)
	}

	var times []time.Time

	// Time to fill in some missing data.
	for _, dataSet := range rawData {
		// Try to short circuit if there's no missing data.
		if len(dataSet.data) == len(fullTimes) {
			continue
		}

		// Otherwise, we're missing values :(

		// Generate a set of all known times from the new time series
		var (
			knownTimes = map[string]struct{}{}
			formatter  = func(t time.Time) string {
				return t.Format("2006:1:2")
			}
		)

		// Generate our known set.
		for _, tim := range dataSet.data {
			knownTimes[formatter(tim.time)] = struct{}{}
		}

		// Find our set difference.
		for timeLabel, timeReal := range fullTimes {
			if _, ok := knownTimes[timeLabel]; ok {
				continue
			}

			dataSet.data = append(dataSet.data, timeValuePair{
				time: timeReal,
				val:  0,
			})
		}

		// Now we need to sort them
		sort.Sort(ByTime(dataSet.data))

		var (
			timePoints []time.Time = make([]time.Time, len(dataSet.data))
			values     []int       = make([]int, len(dataSet.data))
		)

		for i, val := range dataSet.data {
			timePoints[i], values[i] = val.time, val.val
		}

		if times == nil {
			times = timePoints
		}

		dataSeries = append(dataSeries, ChartSeries{
			Name:   dataSet.label,
			Values: values,
		})
	}

	return &Results{
		TimePoints: times,
		Data:       dataSeries,
	}
}

// Results are the analyzed results from our JQL query.
type Results struct {
	TimePoints []time.Time   `json:"timepoints"`
	Data       []ChartSeries `json:"data"`
}
