package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/the42/cartconvert/cartconvert/osgb36"
)

// osGridToMarker provides a wrapper around the functionality from cartconvert/osgb36
func osGridToMarker(name, osGrid string) (Marker, error) {
	coord, err := osgb36.AOSGB36ToStruct(osGrid, osgb36.OSGB36Leave)
	//fmt.Println(coord, osGrid, coord.Easting, coord.Northing)
	if err != nil {
		return Marker{}, err
	}
	latlong := osgb36.OSGB36ToWGS84LatLong(coord)
	fmt.Printf("%1.2f\n", latlong.Longitude)
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
