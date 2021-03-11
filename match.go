// package main finds hostels and waterfalls in the UK which are close to each other.
package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	hostelFile := flag.String("hostels", "hostels.xml", "xml file of hostels with location data")
	waterUrl := flag.String("waterfalls", "https://en.wikipedia.org/wiki/List_of_waterfalls_of_the_United_Kingdom", "waterfalls data url")
	flag.Parse()

	hostels := XmlGetLocations(*hostelFile)
	//fmt.Printf("%v\n", hostels[:5])
	//waterfalls := WikiGetLocations(*waterUrl)
	waterfalls := CrawlWiki(*waterUrl)
	_, _ = waterfalls, hostels
	fmt.Printf("%+v\n", (waterfalls))
}

func MakeWikiUrl(pagename string) string {
	return "https://en.wikipedia.org/wiki/" + pagename
}

func CrawlWiki(listURL string) []Marker {
	lines, err := GetWikiText(listURL)
	if err != nil {
		log.Panic(err)
	}

	var waterfalls []string
	var inSection string = ""
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			inSection = ""
		}
		// We're not going to Northern Ireland this year,
		// and the Scotland waterfall links are messed up. TBC
		if inSection == "Scotland" || inSection == "Northern_Ireland" {
			continue
		}
		if inSection != "" {
			// the data are in a table so the first character is a '*'
			link, err := ParseXWikiLinks(line[1:])
			if err != nil {
				continue
			}
			waterfalls = append(waterfalls, link)
		}
		// country headers are links surrounded by "==="
		if strings.Contains(line, "===") {
			inSection, _ = ParseXWikiLinks(line)
		}
	}

	var formatted []Marker
	for _, f := range waterfalls {
		//fmt.Println(f)
		mark, err := GetLocationFromWikiPage(f)
		if err != nil {
			fmt.Printf("%s: %v\n", f, err)
		} else {
			formatted = append(formatted, mark)
		}
	}

	return formatted
}

func GetWikiText(url string) ([]string, error) {
	page, err := http.Get(url + "?action=raw")
	defer page.Body.Close()
	if page.StatusCode == 404 {
		// when 404, page.Status should be "404 Not Found"
		return []string{""}, errors.New(page.Status)
	}
	pageBytes, _ := ioutil.ReadAll(page.Body)
	lines := strings.Split(fmt.Sprintf("%s", pageBytes), "\n")

	// check if it is a redirect page, and if so, follow it:
	if strings.Contains(lines[0], "REDIRECT") || strings.Contains(lines[0], "redirect") {
		redirectURL, _ := ParseXWikiLinks(lines[0])
		fmt.Printf("\tredirecting to %s\n", redirectURL)
		return GetWikiText(MakeWikiUrl(redirectURL))
	}
	return lines, err
}

func GetLocationFromWikiPage(wikiURL string) (Marker, error) {
	lines, err := GetWikiText(MakeWikiUrl(wikiURL))
	if err != nil {
		return Marker{}, err
	}
	var coord string
	for _, line := range lines {
		if strings.Contains(line, "{{coord") || strings.Contains(line, "{{Coord") {
			// check that there is coord data, not just an empty tag
			if strings.Contains(line, "oord}}") {
				continue
			}
			coord = line
			break
		}
	}
	if coord == "" {
		/*for _, line := range lines {
			fmt.Println(line)
		}*/
		return Marker{}, errors.New("No location found")
	}

	var lat, long float64
	// now we need to extract the location from the coord string, and convert
	// it if necessary to decimal format.
	// firstly, get rid of the "{{coord|" bit (the coordinate starts after it)
	begIndex := strings.Index(coord, "{{coord")
	coord = coord[begIndex+8:]
	// remove any leading "|" s, since they'll mess up the splitting bit
	for coord[0] == '|' {
		coord = coord[1:]
	}
	// locations in the degree, minute, second format contain a "|N|" or "|S|" and "|W" or "|E"
	// (for the W and E it's possible not to have a final pipe since they come last,
	// so if there's no extra info afterwards there might be a }} rather than |.)
	if strings.Contains(coord, "|N|") || strings.Contains(coord, "|S|") {
		var n_or_s string
		if strings.Contains(coord, "|N|") {
			n_or_s = "|N|"
		} else {
			n_or_s = "|S|"
		}
		latS := coord[:strings.Index(coord, n_or_s)]
		lat = DmsToDec(latS)

		var w_or_e string
		if strings.Contains(coord, "|W") {
			w_or_e = "|W"
		} else {
			w_or_e = "|E"
		}
		longS := coord[strings.Index(coord, n_or_s)+3 : strings.Index(coord, w_or_e)]
		long = DmsToDec(longS)

		// correct for west and south coords begin negative in decimal
		if w_or_e == "|W" {
			long = -long
		}
		if n_or_s == "|S|" {
			lat = -lat
		}
	} else {
		// otherwise, the data are already in decimal format
		split_coords := strings.Split(coord, "|")
		lat, _ = strconv.ParseFloat(strings.TrimSpace(split_coords[0]), 64)
		long, _ = strconv.ParseFloat(strings.TrimSpace(split_coords[1]), 64)
	}

	// to make the marker name look nice, change the underscores back to spaces
	name := strings.ReplaceAll(wikiURL, "_", " ")
	return Marker{
		Name: name,
		Lat:  lat,
		Long: long,
	}, nil
}

// DmsToDec takes a string of degrees, minutes, second GPS coords
// which are separated by the character "|"
// and returns the coords in decimal format.
// 3 fields are not required.
func DmsToDec(dms string) float64 {
	s := strings.Split(dms, "|")
	var dec float64 = 0
	for i, _ := range s {
		dms_float, _ := strconv.ParseFloat(s[i], 64)
		dec += dms_float / math.Pow(60.0, float64(i))
	}
	return dec
}

func ParseXWikiLinks(s string) (string, error) {
	if !strings.Contains(s, "[[") && !strings.Contains(s, "]]") {
		return "", errors.New("not a Wiki link")
	}
	// remove leading whitespace or crud
	for s[0] != '[' {
		s = s[1:]
	}
	s = s[2:strings.Index(s, "]]")]
	// remove leading whitespace again
	for s[0] == ' ' {
		s = s[1:]
	}
	// some links are formatted with a part after "|" which is displayed,
	// but with a different target address.
	if strings.Contains(s, "|") {
		s = s[:strings.Index(s, "|")]
	}
	// redirect links can contain links to particular headers with a "#",
	// which can affect the ?action=raw bit to get proper wikitext rather than html
	// so remove the header bit.
	if strings.Contains(s, "#") {
		s = s[:strings.Index(s, "#")]
	}
	// replace all spaces with underscores
	s = strings.Replace(s, " ", "_", -1)

	return s, nil
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

func WikiGetLocations(url string) Wiki {
	page, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}
	defer page.Body.Close()
	fmt.Printf("%s\n", page.Status)
	if page.StatusCode == 404 {
		os.Exit(1)
	}
	//fmt.Printf("http.Get page is a %T\n", page)

	pageBytes, _ := ioutil.ReadAll(page.Body)

	//doc, err := html.Parse(strings.NewReader(fmt.Sprintf("%s", pageBytes)))
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("%+v\n", doc.FirstChild.NextSibling.LastChild.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling)

	var waters Wiki
	if err := xml.Unmarshal([]byte(pageBytes), &waters); err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%v\n", waters.htmls)
	return waters
}

type Markers struct {
	Markers []Marker
}

// Marker is a basic point with a location expressed in decimal coordinates
// and a name.
type Marker struct {
	Name string
	Lat  float64
	Long float64
}

// FindRanges returns the index of the Marker with the largest or smallest
// Lat or Long. The arguments it takes are bools:
// when lat is true, the Lats are searched;
// when lat is false, the Longs are searched;
// when max is true, the maximum is found;
// when max is false, the minimum is found.
func (m Markers) FindRanges(lat bool, max bool) int {
	maxValue := 0.
	var maxIndex int
	var curValue float64
	for i := range m.Markers {
		if lat {
			curValue = m.Markers[i].Lat
		} else {
			curValue = m.Markers[i].Long
		}
		// if what we've got is bigger and we want a bigger one,
		// or if it's smaller and we want a smaller one,
		// log the new value.
		// sidenote: the == here is acting as an XNOR
		if (curValue > maxValue) == max {
			maxValue = curValue
			maxIndex = i
		}
	}
	return maxIndex
}

type Wiki struct {
	html Html `xml:"html"`
}

type IntoDiv struct {
	div Div `xml:"html"`
}

type Html struct {
	XMLName xml.Name `xml:"html"`
	class   string   `xml:",attr"`
	head    string   `xml:"head"`
	body    Body     `xml:"body"`
}

type Body struct {
	div Div `xml:"div"`
}

type Div struct {
	XMLName xml.Name `xml:"html"`
	id      string   `xml:"id,attr"`
	class   string   `xml:"class,attr"`
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
