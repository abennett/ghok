package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	SUMMARY   = "https://www.githubstatus.com/api/v2/components.json"
	INCIDENTS = "https://www.githubstatus.com/api/v2/incidents/unresolved.json"

	NONE     = "✅"
	MINOR    = "🟡"
	MAJOR    = "🟠"
	CRITICAL = "🔴"
	UNKNOWN  = "❔"
)

func toIcon(s string) string {
	switch s := strings.ToLower(s); s {
	case "none", "operational":
		return NONE
	case "minor", "degraded_performance":
		return MINOR
	case "major", "partial_outage":
		return MAJOR
	case "critical", "major_outage":
		return CRITICAL
	default:
		return UNKNOWN
	}
}

type ComponentsResponse struct {
	Page       Page
	Components []Component
}

type IncidentsResponse struct {
	Page      Page
	Incidents []Incident
}

type Page struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Component struct {
	CreatedAt          string  `json:"created_at"`
	Description        string  `json:"description"`
	Group              bool    `json:"group"`
	GroupID            *string `json:"group_id"`
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	OnlyShowIfDegraded bool    `json:"only_show_if_degraded"`
	PageID             string  `json:"page_id"`
	Position           int     `json:"position"`
	Showcase           bool    `json:"showcase"`
	StartDate          string  `json:"start_date"`
	Status             string  `json:"status"`
	UpdatedAt          string  `json:"updated_at"`
}

type Incident struct {
	CreatedAt string   `json:"created_at"`
	ID        string   `json:"id"`
	Impact    string   `json:"impact"`
	Updates   []Update `json:"incident_updates"`
	Name      string   `json:"name"`
	ShortLink string   `json:"shortlink"`
	Status    string   `json:"status"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type Update struct {
	Body       string `json:"body"`
	CreatedAt  string `json:"created_at"`
	DisplayAt  string `json:"display_at"`
	ID         string `json:"id"`
	IncidentID string `json:"incident_id"`
	Status     string `json:"status"`
	UpdatedAt  string `json:"updated_at"`
}

func GetComponents() (*ComponentsResponse, error) {
	resp, err := http.Get(SUMMARY)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(resp.Body)
	var sr ComponentsResponse
	if err = dec.Decode(&sr); err != nil {
		return nil, err
	}
	return &sr, nil
}

func GetIncidents() (*IncidentsResponse, error) {
	resp, err := http.Get(INCIDENTS)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(resp.Body)
	var ir IncidentsResponse
	if err = dec.Decode(&ir); err != nil {
		return nil, err
	}
	return &ir, nil
}

func main() {
	cmpsCh := make(chan *ComponentsResponse, 1)
	incsCh := make(chan *IncidentsResponse, 1)
	go func() {
		cmps, err := GetComponents()
		if err != nil {
			log.Fatal(err)
		}
		cmpsCh <- cmps
	}()
	go func() {
		incs, err := GetIncidents()
		if err != nil {
			log.Fatal(err)
		}
		incsCh <- incs
	}()
	writer := tabwriter.NewWriter(os.Stdout, 4, 4, 1, ' ', 0)
	cmps := <-cmpsCh
	fTime := cmps.Page.UpdatedAt.Format(time.RFC822)
	fmt.Fprintf(writer, "=== Components as of %s === \n", fTime)
	for _, c := range cmps.Components {
		if c.ID == "0l2p9nhqnxpd" {
			continue
		}
		fmt.Fprintf(writer, "%s\t%s\n", c.Name, toIcon(c.Status))
	}
	incs := <-incsCh
	if len(incs.Incidents) > 0 {
		fmt.Fprint(writer, "\n=== Incidents ===")
		for _, i := range incs.Incidents {
            fTime := i.UpdatedAt.Format(time.RFC822)
			fmt.Fprintf(writer, "\nName:\t%s\n", i.Name)
			fmt.Fprintf(writer, "Impact:\t%s %s\n", toIcon(i.Impact), i.Impact)
			fmt.Fprintf(writer, "Status:\t%s\n", i.Status)
			fmt.Fprintf(writer, "Details:\t%s\n", i.Updates[0].Body)
			fmt.Fprintf(writer, "Link:\t%s\n", i.ShortLink)
            fmt.Fprintf(writer, "Last Updated:\t%s\n", fTime)
		}
	}
	writer.Flush()
}
