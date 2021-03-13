package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type mapboxDetails struct {
	uname  string
	style  string
	apikey string
}

// MapboxStatic creates a static image of a map
// with markers represented by m
func MapboxStatic(m Markers, fname string, mbox mapboxDetails) error {
	baseURL := "https://api.mapbox.com/"
	query := fmt.Sprintf("styles/v1/%s/%s/static/", mbox.uname, mbox.style)
	// it is possible to just use "auto" instead of a bbox for this next field
	// (which defines the field of view)
	// and then the overlays (markers) are used to fit the field of view.
	suffix := "/" + formatBounds(m, 0.05) /* "auto"*/ + "/800x920?access_token="
	if *verbose {
		fmt.Printf("Bounds for %s: %v\n", fname, formatBounds(m, 0.05))
	}

	var markersMapbox string

	for _, mark := range m.Markers {
		markersMapbox += markerToMapbox(mark, "", "") + ","
	}
	// remove the final comma
	markersMapbox = markersMapbox[:len(markersMapbox)-1]

	imgP, err := http.Get(baseURL + query + markersMapbox + suffix + mbox.apikey)
	//fmt.Printf("GET: %s\n", baseURL+query+markersMapbox+suffix+mbox.apikey)
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
// with optional spacing as a fraction of the width/height.
func formatBounds(m Markers, space float64) string {
	left := m.Markers[m.FindRanges(false, false)].Long
	bot := m.Markers[m.FindRanges(true, false)].Lat
	right := m.Markers[m.FindRanges(false, true)].Long
	top := m.Markers[m.FindRanges(true, true)].Lat
	lrspace := (right - left) * space
	btspace := (top - bot) * space
	return fmt.Sprintf("[%f,%f,%f,%f]", left-lrspace, bot-btspace, right+lrspace, top+btspace)
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
