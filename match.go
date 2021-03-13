// finds hostels and waterfalls in the UK which are close to each other.
package main

// BUG(): Markers.FindRanges() doesn't seem to work for the waterfall data.

import (
	"encoding/csv"
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
	"time"
)

var (
	verbose *bool
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\t[-v] [-h]\n"+
		"\t\t\t[-hostelFile hostels.xml] [-waterfallsURL https://en.wikipedia.org/wiki/List...]\n"+
		"\t\t\t[-use-cache] [-hostelCache hostels_cache.csv] [-waterfallCache waterfalls_cache.csv]\n"+
		"\t\t\t[-static]\n"+
		"\t\t\t[-sqluname username] [-sqlpwd password] [-sqldb myDB]\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nholiday-plan is a program to get, save, or plot data about hostels and waterfalls in the UK.\n"+
		"If a SQL username is provided, the SQL database is either written to using the obtained data, or read from, if use-cache is true.\n")
}

func main() {
	hostelFile := flag.String("hostelFile", "hostels.xml", "xml file of hostels with location data")
	waterURL := flag.String("waterfallsURL", "https://en.wikipedia.org/wiki/List_of_waterfalls_of_the_United_Kingdom", "waterfalls data url")
	verbose = flag.Bool("v", false, "print verbose output to stderr")
	hostelSave := flag.String("hostelCache", "hostels_cache.csv", "saves hostel data to the file")
	waterfallSave := flag.String("waterfallCache", "waterfalls_cache.csv", "saves waterfall data to the file")
	useCache := flag.Bool("use-cache", false, "use the cache rather than File/URL (requires the cache filename flags)")

	staticImgs := flag.Bool("static", false, "generate static PNGs of maps with markers")
	var mboxDs mapboxDetails
	flag.StringVar(&mboxDs.uname, "mapboxuname", "", "mapbox.com username")
	flag.StringVar(&mboxDs.style, "mapboxstyle", "", "style of mapbox map")
	flag.StringVar(&mboxDs.apikey, "mapboxapi", "", "api key for mapboxuname")

	flag.Usage = usage
	flag.Parse()

	if *staticImgs && (mboxDs.uname == "" || mboxDs.style == "" || mboxDs.apikey == "") {
		log.Fatal("insufficient credentials provided to generate mapbox maps")
	}

	var hostels, waterfalls Markers
	var err error

	if *useCache {
		if *waterfallSave == "" {
			log.Fatal("Please provide the filename of the waterfall cache")
		}
		if *hostelSave == "" {
			log.Fatal("Please provide the filename of the hostel cache")
		}
		hostels, err = CSVtoMarkers(*hostelSave)
		if err != nil {
			log.Fatal(err)
		}
		waterfalls, err = CSVtoMarkers(*waterfallSave)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Reading hostels XML...\n")
		hostels = KMLGetLocations(*hostelFile)
		fmt.Fprintf(os.Stderr, "Crawling waterfalls list webpage...\n")
		waterfalls = crawlWiki(*waterURL)

		// save cached data to file
		n, err := hostels.SaveCSV(*hostelSave)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Fprintf(os.Stderr, "saved %d bytes to %s\n", n, *hostelSave)
		}
		n, err = waterfalls.SaveCSV(*waterfallSave)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Fprintf(os.Stderr, "saved %d bytes to %s\n", n, *waterfallSave)
		}
	}
	// not sure how useful this is
	if *verbose {
		for _, m := range hostels.Markers {
			fmt.Printf("%v\n", m)
		}
		fmt.Println()
		for _, m := range waterfalls.Markers {
			fmt.Printf("%v\n", m)
		}
	}
	fmt.Fprintf(os.Stderr, "Got %v hostels, %v waterfalls\n", len(hostels.Markers), len(waterfalls.Markers))

	if *staticImgs {
		hostelsImg := "map-hostels-" + time.Now().Format("2006-01-02-1504") + ".png"
		err = MapboxStatic(hostels, hostelsImg, mboxDs)
		if err != nil {
			log.Fatal(err)
		}

		waterfallsImg := "map-waterfalls-" + time.Now().Format("2006-01-02-1504") + ".png"
		err = MapboxStatic(waterfalls, waterfallsImg, mboxDs)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Wrote maps to %s and %s.\n", hostelsImg, waterfallsImg)
	}

}

// longestName returns the index and length of the longest Name in a Markers.
func (m Markers) longestName() (index int, length int) {
	length = 0
	index = 0
	for i := range m.Markers {
		if len(m.Markers[i].Name) > length {
			length = len(m.Markers[i].Name)
			index = i
		}
	}
	return
}

// CSVtoMarkers takes the name of a CSV file and returns a Markers.
// The CSV may have been produced by Markers.SaveCSV:
// the file has three fields: name, lat, long.
func CSVtoMarkers(fname string) (Markers, error) {
	f, err := os.Open(fname)
	if err != nil {
		return Markers{}, err
	}
	r := csv.NewReader(f)
	lines, err := r.ReadAll()
	if err != nil {
		return Markers{}, err
	}
	m := Markers{make([]Marker, len(lines))}
	for i, line := range lines {
		m.Markers[i].Name = line[0]
		m.Markers[i].Lat, err = strconv.ParseFloat(line[1], 64)
		m.Markers[i].Long, _ = strconv.ParseFloat(line[2], 64)
	}

	return m, err
}

// MakeWikiURL takes a formatted Wikipedia pagename (ie spaces are underscores)
// and adds the English Wikipedia prefix.
func MakeWikiURL(pagename string) string {
	return "https://en.wikipedia.org/wiki/" + pagename
}

func crawlWiki(listURL string) Markers {
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

	fmt.Fprintf(os.Stderr, "Parsed list page, following links...\n")

	var formatted Markers
	for _, f := range waterfalls {
		//fmt.Println(f)
		mark, err := GetLocationFromWikiPage(f)
		if err != nil {
			if *verbose {
				fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			}
		} else {
			formatted.Markers = append(formatted.Markers, mark)
		}
	}

	return formatted
}

// GetWikiText takes the url of a normal Wikipedia page
// and returns the lines in text/x-wiki format.
// It follows redirects and says so by printing to Stderr.
func GetWikiText(url string) ([]string, error) {
	page, err := http.Get(url + "?action=raw")
	if err != nil {
		return []string{""}, err
	}
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
		if *verbose {
			fmt.Fprintf(os.Stderr, "%s :  redirecting to %s\n", url, redirectURL)
		}
		return GetWikiText(MakeWikiURL(redirectURL))
	}
	return lines, err
}

// GetLocationFromWikiPage takes a Wikipedia page name, and returns
// a Marker with the pagename as a Name and the location from a {{coord}}
// tag in the Wikitext.
// The location data is converted to decimal form if necessary.
func GetLocationFromWikiPage(wikiURL string) (Marker, error) {
	lines, err := GetWikiText(MakeWikiURL(wikiURL))
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
	begIndex := strings.Index(coord, "{{coord|")
	if begIndex == -1 {
		begIndex = strings.Index(coord, "{{Coord|")
	}
	// The + 1 in these next two account for the extra space.
	// Really, any amount of whitespace should be checked for.
	if begIndex == -1 {
		begIndex = strings.Index(coord, "{{coord |") + 1
	}
	if begIndex == -1 {
		begIndex = strings.Index(coord, "{{Coord |") + 1
	}
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
	for i := range s {
		dms_float, _ := strconv.ParseFloat(s[i], 64)
		dec += dms_float / math.Pow(60.0, float64(i))
	}
	return dec
}

// ParseXWikiLinks takes a string and returns a correctly formatted
// string of the Wikipedia url suffix of the first link found in the string.
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

// KMLGetLocations extracts location data from an xml file.
func KMLGetLocations(filename string) Markers {
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
	formatted := make([]Marker, n)
	for i, place := range hostels.Documents.Folders[0].Placemarks {
		formatted[i].Name = place.Name
		gps := strings.Split(place.Point.Coords, ",")
		formatted[i].Long, _ = strconv.ParseFloat(gps[0], 64)
		formatted[i].Lat, _ = strconv.ParseFloat(gps[1], 64)
	}

	return Markers{Markers: formatted}
}

// Markers wraps a slice of type Marker
type Markers struct {
	Markers []Marker
}

// Marker is a basic point with a name and a location expressed in decimal coordinates
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
	var maxValue float64
	if lat {
		maxValue = m.Markers[0].Lat
	} else {
		maxValue = m.Markers[1].Long
	}
	var maxIndex int = 0
	var curValue float64 = 0.0
	for i := range m.Markers[1:] {
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

// SaveCSV saves a Markers to a file in CSV format,
// with each line of the CSV having 3 fields, delimited by commas,
// representing Markers[i].Name, Lat, Long respectively.
// The first field (Name) is surrounded by double quotes (" ").
// The number of bytes written and an error is returned.
// if the filename provided already exists, an error is returned.
func (m Markers) SaveCSV(filename string) (int, error) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return 0, err
	}

	var bytesWritten int
	for _, mark := range m.Markers {
		n, err := fmt.Fprintln(f, fmt.Sprintf("\"%s\",%f,%f", mark.Name, mark.Lat, mark.Long))
		if err != nil {
			return bytesWritten, nil
		}
		bytesWritten += n
	}

	return bytesWritten, err
}

// Kml provides the highest level of tags in a KML-type XML file
// which are relevant for finding location data.
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
