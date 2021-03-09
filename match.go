package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func main() {
	hostelFile := flag.String("hostels", "hostels.xml", "xml file of hostels with location data")
	waterUrl := flag.String("waterfalls", "https://en.wikipedia.org/wiki/List_of_waterfalls_of_the_United_Kingdom", "waterfalls data url")
	flag.Parse()

	hostels := XmlGetLocations(*hostelFile)
	//fmt.Printf("%v\n", hostels[:5])
	waterfalls := WikiGetLocations(*waterUrl)
	_, _ = waterfalls, hostels
	_ = hostels
	//fmt.Printf("%T\n", (hostels))
}

func XmlGetLocations(filename string) []Marker {
	hFile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer hFile.Close()

	hostelBytes, _ := ioutil.ReadAll(hFile)

	var hostels Kml
	if err := xml.Unmarshal([]byte(hostelBytes), &hostels); err != nil {
		log.Fatal(err)
	}

	// the first element in Folders in this case is YHA hostels.
	// Other elements are either retired or independent, and negligible.

	n := len(hostels.Documents.Folders[0].Placemarks)
	fmt.Println("n: ", n)
	formatted := make([]Marker, n)
	for i, place := range hostels.Documents.Folders[0].Placemarks {
		formatted[i].Name = place.Name
		gps := strings.Split(place.Point.Coords, ",")
		formatted[i].Long, _ = strconv.ParseFloat(gps[0], 64)
		formatted[i].Lat, _ = strconv.ParseFloat(gps[1], 64)
	}

	return formatted
}

func WikiGetLocations(url string) error {
	page, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer page.Body.Close()
	fmt.Printf("%s\n", page.Status)
	if page.StatusCode == 404 {
		os.Exit(1)
	}
	fmt.Printf("http.Get page is a %T\n", page)

	pageBytes, _ := ioutil.ReadAll(page.Body)

	doc, err := html.Parse(strings.NewReader(fmt.Sprintf("%s", pageBytes)))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", doc.FirstChild.NextSibling.LastChild.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling)

	var waters Wiki
	if err := xml.Unmarshal([]byte(pageBytes), &waters); err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%v\n", waters.htmls)
	return err
}

type Marker struct {
	Name string
	Lat  float64
	Long float64
}

type Wiki struct {
	html Html `xml:"html"`
}

type Html struct {
	head string `xml:"head"`
	body Body   `xml:"body"`
}

type Body struct {
	div Div `xml:"div"`
}

type Div struct {
	id    string `xml:"id,attr"`
	class string `xml:"class,attr"`
	//div   Div    `xml:"div"`
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
