/* Copyright 2021 Ben Fuller
 * Apache License, Version 2.0
 * See LICENCE file for copyright and licence details.
 */

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func saveMapboxHTML(fname, js string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	var html string = `<!DOCTYPE html><html><head><meta charset="utf-8" /><meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1, user-scalable=no" />
<title>Map: plan for holiday</title>
<!--
This work by Ben Fuller is licensed under CC-BY-SA 4.0:
http://creativecommons.org/licenses/by-sa/4.0/
-->
<!-- data source for markers are the Wikipedia articles
"List of Waterfalls of the United Kingdom" and "List of Youth Hostels in England and Wales"
which are licensed under CC-BY-SA 3.0. -->
<!-- favicon source:
Copyright 2020 Twitter, Inc and other contributors (https://github.com/twitter/twemoji)
https://github.com/twitter/twemoji/blob/master/assets/svg/1f9e1.svg
License: CC-BY 4.0 -->
<link rel="apple-touch-icon" sizes="180x180" href="apple-touch-icon.png">
<link rel="icon" type="image/png" sizes="32x32" href="favicon-32x32.png">
<link rel="icon" type="image/png" sizes="16x16" href="favicon-16x16.png">
<link href="https://api.mapbox.com/mapbox-gl-js/v2.1.1/mapbox-gl.css" rel="stylesheet"> <script src="https://api.mapbox.com/mapbox-gl-js/v2.1.1/mapbox-gl.js"></script>
<style>
	body {
		margin: 0;
		padding: 0;
	}
	#map {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 100%;
	}
	#menu {
		position: absolute;
		background: rgba(239, 239, 239, 0.3);
		padding: 10px;
		font-family: sans-serif;
	}
</style>
</head>
<body>
<div id="map"></div>
<div id="menu">
<input id="outdoors-v11" type="radio" name="rtoggle" value="outdoors" checked="checked">
<label for="outdoors-v11">outdoors</label>
<input id="satellite-v9" type="radio" name="rtoggle" value="satellite">
<label for="satellite-v9">satellite</label>
<input id="streets-v11" type="radio" name="rtoggle" value="streets">
<label for="streets-v11">streets</label>
</div>
<script>
` + js + `
</script>
</body>
</html>
`

	_, err = f.Write([]byte(html))
	return err
}

func mapboxEmbeddedPage(fname, mapURL, content string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	var html string = ` <!DOCTYPE html><html><head><meta charset="utf-8" /> <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=yes" />
<title>Plan for holiday</title>
<!-- data sources for table are the Wikipedia articles
"List of Waterfalls of the United Kingdom" and "List of Youth Hostels in England and Wales"
which are licensed under CC-BY-SA 3.0. -->
<!-- favicon source:
Copyright 2020 Twitter, Inc and other contributors (https://github.com/twitter/twemoji)
https://github.com/twitter/twemoji/blob/master/assets/svg/1f9e1.svg
License: CC-BY 4.0 -->
<link rel="apple-touch-icon" sizes="180x180" href="apple-touch-icon.png">
<link rel="icon" type="image/png" sizes="32x32" href="favicon-32x32.png">
<link rel="icon" type="image/png" sizes="16x16" href="favicon-16x16.png">
<link href="https://api.mapbox.com/mapbox-gl-js/v2.1.1/mapbox-gl.css" rel="stylesheet"> <script src="https://api.mapbox.com/mapbox-gl-js/v2.1.1/mapbox-gl.js"></script>
<style>
	body{
		margin:1em auto;
		margin-top: 100px;
		max-width: 40em;
		line-height: 1.4;
		font: 1.2em/1.62 sans-serif;
		padding: 0 0.62em;
		color: #444;
		background: #eeeeee;
	}
	h1{
		text-align: center;
		line-height: 1.2;
	}
	h2,h4{
		text-align:right;
		line-height: 1.2;
	}
	table, th, td {
		border: 1px solid black;
		border-collapse: collapse;
	}
	th, td {
		padding: 15px;
	}
	footer {
		font-size: 0.6em;
		padding: 5px;
		border-top: 1px solid black;
	}
	.right {
		float: right;
	}
	.left {
		float: left;
	}
	@media(prefers-color-scheme:dark) {
		body{
			background: #292929;
			color: #fff;
		}
		table, th, td {
			color: #fff;
		}
		a {
			color: #6cf;
		}
	}
	@media(prefers-color-scheme: light){
		body{
			background: #eeeeee;
			color: #444;
		}
		table, th, td {
			color: #444;
		}
	}
</style>
</head>
<body>
<a id="top"></a>
<h1>Map of Waterfalls in the UK</h1>
<h2>And hostels close to them</h2>
<p>
The map below shows waterfalls with blue markers and hostels with brown markers.

All the YHA hostels in England and Wales are shown, but the hostels which are closest to a waterfall are larger.

Click on any marker to show its name, and a link for the hostels.

Underneath there is a table showing for each waterfall which hostel is nearest, again with links to YHA hostel pages.
</p>
<center>
<iframe name="map" id="map" allowfullscreen="" src=` + fmt.Sprintf("%q", mapURL) + ` height="500" width="500" style="max-width:100%;"></iframe>
</center>
<p>You can see a fullscreen version of this map <a href=` + fmt.Sprintf("%q", mapURL) + `>here</a>.</p>
` + content + `

<br>
<footer>
<div class="left" id="license">
<a rel="license" href="http://creativecommons.org/licenses/by-sa/4.0/"><img alt="Creative Commons BY-SA 4.0 Licence" style="border-width:0" src="https://i.creativecommons.org/l/by-sa/4.0/80x15.png" /></a>
<br />
Copyright © Ben Fuller, 2021
<br>
<br>
</div>
<div class="right"><a href="#top">↑ Back to top</a></div>
</footer>
</body>
</html>
`

	_, err = f.Write([]byte(html))
	return err
}

// mapToTable takes a map[string][]string and turns it into a html table
// sorted by the key.
// headers must have two elements; one to head the keys and one the values.
func mapToTable(m map[string][]string, headers ...interface{}) string {
	// sort the keys
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var tableBody string
	for _, k := range keys {
		if len(m[k]) == 0 {
			continue
		}
		tableBody += fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>\n", k, strings.Join(m[k], "<br>"))
	}

	return fmt.Sprintf("<table id=\"table\">\n<tr><th>%s</th><th>%s</th></tr>\n", headers...) + tableBody + "</table>"
}

// assumes a map variable called map in the rest of the js
func markerToJS(m Markers, color string) string {
	var js string
	markerTemplate := "new mapboxgl.Marker({color: %q, scale: %f}).setLngLat([%f,%f]).setPopup(new mapboxgl.Popup({offset: 25}).setText(%q)).addTo(map);\n"

	for _, mark := range m.Markers {
		js = js + fmt.Sprintf(markerTemplate, color, mark.scale, mark.Long, mark.Lat, mark.Name)
	}

	return js
}

func mapboxMapJS(mbox mapboxDetails, bbox string) string {
	return `mapboxgl.accessToken = ` + fmt.Sprintf("%q", mbox.apikey) + `;
var bbox = ` + bbox + `;
var map = new mapboxgl.Map({
	container: 'map',
	style:` + fmt.Sprintf("%q", mbox.style) + `,
});
map.fitBounds(bbox);

var layerList = document.getElementById('menu');
var inputs = layerList.getElementsByTagName('input');
 
function switchLayer(layer) {
var layerId = layer.target.id;
map.setStyle('mapbox://styles/mapbox/' + layerId);
}
 
for (var i = 0; i < inputs.length; i++) {
inputs[i].onclick = switchLayer;
}
`
}
