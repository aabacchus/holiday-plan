package main

import (
	"fmt"
	"os"
)

func saveMapboxHTML(fname, js string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	var html string = `<!DOCTYPE html><html><head><meta charset="utf-8" /><meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1, user-scalable=no" />
<title>Map: plan for holiday</title>
<!-- favicon source:
Copyright 2020 Twitter, Inc and other contributors (https://github.com/twitter/twemoji)
https://github.com/twitter/twemoji/blob/master/assets/svg/1f9e1.svg
License: CC-BY 4.0 -->
<link rel="icon" href="favicon.ico" />
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
<!-- favicon source:
Copyright 2020 Twitter, Inc and other contributors (https://github.com/twitter/twemoji)
https://github.com/twitter/twemoji/blob/master/assets/svg/1f9e1.svg
License: CC-BY 4.0 -->
<link rel="icon" href="favicon.ico" />
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
<iframe id="map" allowfullscreen="" src=` + fmt.Sprintf("%q", mapURL) + ` height="500" width="500"></iframe>
` + content + `

</body>
</html>
`

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
