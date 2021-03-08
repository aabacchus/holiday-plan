package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	hostelFile := flag.String("hostels", "hostels.xml", "xml file of hostels with location data")
	//waterFile := flag.String("waterfalls", "waterfalls.xml", "waterfalls file")
	flag.Parse()

	hFile, err := os.Open(*hostelFile)
	if err != nil {
		log.Fatal(err)
	}
	defer hFile.Close()

	hostelBytes, _ := ioutil.ReadAll(hFile)

	//fmt.Printf("%s", hostelBytes)

	var hostels Kml

	if err := xml.Unmarshal([]byte(hostelBytes), &hostels); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", (hostels.Documents.Folders[0].Placemarks))
}

type Kml struct {
	XMLName   xml.Name `xml:"kml"`
	Documents Document `xml:"Document"`
}

type Document struct {
	XMLName xml.Name `xml:"Document"`
	Folders []Folder `xml:"Folder"`
}

type Folder struct {
	XMLName    xml.Name    `xml:"Folder"`
	Name       string      `xml:"name"`
	Placemarks []Placemark `xml:"Placemark"`
}

type Placemark struct {
	Name  string `xml:"name"`
	Point Coord  `xml:"Point"`
}

type Coord struct {
	Coords string `xml:"coordinates"`
}
