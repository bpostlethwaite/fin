package finpony

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

func PlaceTypes(query string) ([][]string, error) {
	c, err := maps.NewClient(maps.WithAPIKey(ConfigCreds().GoogleMapsApiKey))
	if err != nil {
		return nil, err
	}

	tSearch := maps.TextSearchRequest{
		Query: query,
	}
	resp, err := c.TextSearch(context.Background(), &tSearch)
	if err != nil {
		return nil, err
	}

	cats := [][]string{}
	for _, r := range resp.Results {
		cats = append(cats, r.Types)
	}

	return cats, nil
}
