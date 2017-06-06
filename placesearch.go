package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"googlemaps.github.io/maps"
)

const TEXT_URL = "https://maps.googleapis.com/maps/api/place/textsearch/json"

func queryUrl(query string) string {
	return fmt.Sprintf("%s?query=%s&key=%s",
		TEXT_URL, query, ConfigCreds().GoogleMapsApiKey)
}

func PlaceType(query string) (string, error) {
	c, err := maps.NewClient(maps.WithAPIKey(ConfigCreds().GoogleMapsApiKey))
	if err != nil {
		return "", err
	}

	tSearch := maps.TextSearchRequest{
		Query: query,
	}
	resp, err := c.TextSearch(context.Background(), &tSearch)
	if err != nil {
		return "", err
	}

	var category string
	if len(resp.Results) > 0 {
		if len(resp.Results[0].Types) > 0 {
			category = resp.Results[0].Types[0]
		}
	}

	return category, nil
}

var replDigits *regexp.Regexp = regexp.MustCompile("(.*\\s+)([0-9]+(-?[0-9]+)+)(\\s+.*)")

func GooglePlaces(txs []Record, cats [][]string) ([]Record, error) {

	recs := []Record{}
	// remove weird digits and whatnot
	for i, _ := range txs {
		if txs[i].Category != "" {
			continue
		}
		subName := replDigits.ReplaceAllString(txs[i].Name, "$1$4")
		gcat, err := PlaceType(subName)
		if err != nil {
			if strings.Contains(err.Error(), "ZERO_RESULTS") {
				continue
			}
			return nil, err
		}

		cat := getCategoryFromRelated(gcat, cats)
		if cat != "" {
			txs[i].Category = cat
			recs = append(recs, txs[i])
		}
	}

	return recs, nil
}

func getCategoryFromRelated(rel string, cats [][]string) string {
	lrel := strings.ToLower(rel)
	for _, row := range cats {
		cat := row[0]
		for _, v := range row {
			if strings.ToLower(v) == lrel {
				return cat
			}
		}
	}

	return ""
}
