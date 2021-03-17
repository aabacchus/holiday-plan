package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/the42/cartconvert/cartconvert/osgb36"
)

// jsonToMarkers takes json input and returns all the location infos with a location.
func jsonToMarkers(jsons []byte) (Markers, error) {
	var out struct{ Data []hostelJSONPlace }
	err := json.Unmarshal(jsons, &out)
	var hostels Markers
	for _, p := range out.Data {
		var lat, long float64
		if p.Lat == "" || p.Lng == "" {
			continue
		}
		lat, err = strconv.ParseFloat(p.Lat, 64)
		long, _ = strconv.ParseFloat(p.Lng, 64)
		if err != nil {
			return hostels, err
		}
		hostels.Markers = append(hostels.Markers, Marker{
			Name: p.Name,
			Lat:  lat,
			Long: long,
		})
	}
	return hostels, err
}

type hostelJSONPlace struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Lng  string `json:"lng"`
}

func scottishHostels(jsonURL string) (Markers, error) {
	var hostels Markers
	f, err := http.Get(jsonURL)
	if err != nil {
		return hostels, err
	}
	defer f.Body.Close()
	bytes, _ := ioutil.ReadAll(f.Body)
	hostels, err = jsonToMarkers(bytes)
	return hostels, err
}

// osGridToMarker provides a wrapper around the functionality from cartconvert/osgb36
func osGridToMarker(name, osGrid string) (Marker, error) {
	coord, err := osgb36.AOSGB36ToStruct(osGrid, osgb36.OSGB36Leave)
	//fmt.Println(coord, osGrid, coord.Easting, coord.Northing)
	if err != nil {
		return Marker{}, err
	}
	latlong := osgb36.OSGB36ToWGS84LatLong(coord)
	return Marker{
		Name:  name,
		Lat:   latlong.Latitude,
		Long:  latlong.Longitude,
		scale: 0.8,
	}, nil
}

func wikiScotlandParse() (Markers, error) {
	var waterfalls Markers

	// download list
	lines, err := GetWikiText("https://en.wikipedia.org/wiki/List_of_waterfalls_of_Scotland")
	if err != nil {
		return waterfalls, err
	}

	// parse lines of list
	var inLocation bool = false
	var lineInLoc int = 0
	var name, gridref string
	for _, line := range lines {
		if line == "" {
			inLocation = false
			continue
		}
		// each location bit in the tables is formatted like this:
		// |-
		// |[[Name of Waterfall]]
		// |[[River]]
		// |{{gbm4ibx|gridref}}
		// |[[general area]]
		// and then the next one starts, or that section ends.
		// We only need the name and grid reference from each place.
		if inLocation {
			lineInLoc++
			if lineInLoc == 1 {
				name, err = ParseXWikiLinks(line)
				if err != nil {
					// if the name isn't made into a link
					// get rid of the first "|" and remove any whitespace
					name = line[1:]
					name = strings.TrimSpace(name)
				}
				name = strings.ReplaceAll(name, "_", " ")
			}
			if lineInLoc == 3 {
				idx := strings.Index(line, "{{gbm4ibx|")
				if idx == -1 {
					log.Printf("%s: no loc found\n%s", name, line)
				}
				endidx := strings.Index(line, "}}")
				if idx == -1 {
					log.Printf("%s: no end of loc found\n%s", name, line)
				}
				gridref = line[idx+10 : endidx]
				mark, err := osGridToMarker(name, gridref)
				if err != nil {
					log.Println("error parsing grid reference")
					log.Panic(err)
				}
				waterfalls.Markers = append(waterfalls.Markers, mark)
				inLocation = false
			}
		}
		if line[:2] == "|-" {
			inLocation = true
			lineInLoc = 0
		}
	}

	return waterfalls, nil
}
