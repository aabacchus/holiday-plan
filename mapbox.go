package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// MapboxStatic creates a static image of a map
// with markers represented by m
func MapboxStatic(m Markers, fname string) error {
	baseURL := "https://api.mapbox.com/"
	query := "styles/v1/phoebos/ckm5olymje9lp17qnq8108oro/static/"
	// it is possible to just use "auto" instead of a bbox for this next field
	// (which defines the field of view)
	// and then the overlays (markers) are used to fit the field of view.
	suffix := "/" + /* formatBounds(m) */ "auto" + "/800x920?access_token="
	api_key := "sk.eyJ1IjoicGhvZWJvcyIsImEiOiJja201b2o4ZHEwZzl6Mm5ud2o4bXd0NjhlIn0.lEiKI4kRVe0Ao4TConJfQQ"

	var markersMapbox string

	for _, mark := range m.Markers {
		markersMapbox += markerToMapbox(mark, "", "") + ","
	}
	// remove the final comma
	markersMapbox = markersMapbox[:len(markersMapbox)-1]

	imgP, err := http.Get(baseURL + query + markersMapbox + suffix + api_key)
	fmt.Printf("GET: %s\n", baseURL+query+markersMapbox+suffix+api_key)
	if err != nil {
		return err
	}
	defer imgP.Body.Close()
	bytes, _ := ioutil.ReadAll(imgP.Body)
	f, err := os.Create(fname)
	f.Write(bytes)
	return err
}

// formatBounds makes a bbox given a Markers
// in the order min(long), min(lat), max(long), max(lat)
func formatBounds(m Markers) string {
	return fmt.Sprintf("[%f,%f,%f,%f]", m.Markers[m.FindRanges(false, false)].Long, m.Markers[m.FindRanges(true, false)].Lat, m.Markers[m.FindRanges(false, true)].Long, m.Markers[m.FindRanges(true, true)].Lat)
}

// markerToMapbox takes a Marker which has a position, and optionally a label and color
// (if you don't want these, provide empty strings)
// and returns the correctly formatted marker for Mapbox
func markerToMapbox(m Marker, label string, color string) string {
	if label != "" {
		label = "-" + label
	}
	if color != "" {
		color = "+" + label
	}
	return fmt.Sprintf("pin-s%s%s(%f,%f)", label, color, m.Long, m.Lat)
}
