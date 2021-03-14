package main

import (
	"fmt"
	"os"
)

func saveMapboxHTML(fname, js, content string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}

	var html string = `<!DOCTYPE html>
	<html>
	<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=yes" />
	<title>Plan for holiday</title>
	<link href="https://api.mapbox.com/mapbox-gl-js/v2.1.1/mapbox-gl.css" rel="stylesheet">
	<script src="https://api.mapbox.com/mapbox-gl-js/v2.1.1/mapbox-gl.js"></script>
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
		#map {
			position: absolute;
			top: 0;
			bottom: 0;
			width: 100%;
		}
		h1{
			text-align: center;
			line-height: 1.2;
		}
		h2,h4{
			text-align:right;
			line-height: 1.2;
		}
		@media(prefers-color-scheme:dark) {
			body{
				background: #292929;
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
		}
	</style>
	</head>
	<body>
	<div id="map"></div>
	` + content + `
	<script>
	` + js + `
	</script>
	</body>
	</html>`

	_, err = f.Write([]byte(html))
	return err
}

// assumes a map variable called map in the rest of the js
func markerToJS(m Markers, color string) string {
	var js string
	markerTemplate := "new mapboxgl.Marker({color: %q}).setLngLat([%f,%f]).addTo(map);\n"

	for _, mark := range m.Markers {
		js = js + fmt.Sprintf(markerTemplate, color, mark.Long, mark.Lat)
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

`
}
