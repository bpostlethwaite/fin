package main

import (
	"context"
	"fmt"

	"googlemaps.github.io/maps"
)

const TEXT_URL = "https://maps.googleapis.com/maps/api/place/textsearch/json"

func queryUrl(query string) string {
	return fmt.Sprintf("%s?query=%s&key=%s",
		TEXT_URL, query, ConfigCreds().GoogleMapsApiKey)
}

func PlaceTypes(query string) (string, error) {
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
