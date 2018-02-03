package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"google.golang.org/api/googleapi/transport"
)

func geocodeUrl(location string, apiKey string) string {
	return fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s", location, apiKey)
}

func geocodeToBytes(location string, apiKey string) []byte {
	url := geocodeUrl(location, apiKey)
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return body
}

func geocode(location string, apiKey string) GeocodingResponse {
	resBytes := geocodeToBytes(location, apiKey)
	fmt.Println(string(resBytes))

	var res GeocodingResponse
	err := json.Unmarshal(resBytes, &res)

	if err != nil {
		panic(err)
	}
	return res

}

type GeocodingResult struct {
	AddressComponents []AddressComponent `json:"address_components"`
	FormattedAddress  string             `json:"formatted_address"`
	Geometry          AddressGeometry    `json:"geometry"`
	Types             []string           `json:"types"`
	PlaceID           string             `json:"place_id"`
}

// AddressComponent is a part of an address
type AddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

// AddressGeometry is the location of a an address
type AddressGeometry struct {
	Location     LatLng       `json:"location"`
	LocationType string       `json:"location_type"`
	Bounds       LatLngBounds `json:"bounds"`
	Viewport     LatLngBounds `json:"viewport"`
	Types        []string     `json:"types"`
}
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
type LatLngBounds struct {
	NorthEast LatLng `json:"northeast"`
	SouthWest LatLng `json:"southwest"`
}

type GeocodingResponse struct {
	Results []GeocodingResult `json:"results"`
}
