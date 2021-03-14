package main

import (
	"math"
)

// matchClosest matches each child to its closest node
func matchClosest(childs, nodes Markers) map[string][]string {
	var matched = make(map[string][]string)
	for _, node := range nodes.Markers {
		matched[node.Name] = []string{}
	}
	for _, child := range childs.Markers {
		closestNode := nodes.sortClosest(child).Name
		matched[closestNode] = append(matched[closestNode], child.Name)
	}
	return matched
}

// sortClosest returns the nearest m.Marker to n
func (m Markers) sortClosest(n Marker) Marker {
	var closest Marker = m.Markers[0]
	var mdist float64 = distanceBn(m.Markers[0], n)
	for _, mark := range m.Markers {
		if d := distanceBn(mark, n); d < mdist {
			mdist = d
			closest = mark
		}
	}
	return closest
}

func distanceBn(m1, m2 Marker) float64 {
	// in meters
	earthR := 6371.009e3
	// convert to radians
	lat1 := m1.Lat * math.Pi / 180
	long1 := m1.Long * math.Pi / 180
	lat2 := m2.Lat * math.Pi / 180
	long2 := m2.Long * math.Pi / 180
	// special case of the Vicenty formula; obtained from https://en.wikipedia.org/wiki/Great-circle_distance
	deltaLong := long2 - long1
	centralAng := math.Atan(
		math.Sqrt(
			math.Pow(math.Cos(lat2)*math.Sin(deltaLong), 2)+
				math.Pow(math.Cos(lat1)*math.Sin(lat2)-
					math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLong), 2)) /
			(math.Sin(lat1)*math.Sin(lat2) +
				math.Cos(lat1)*math.Cos(lat2)*math.Cos(deltaLong)))
	return earthR * centralAng
}

func simplerDistanceBn(m1, m2 Marker) float64 {
	// in meters
	earthR := 6371.009e3
	// convert to radians
	lat1 := m1.Lat * math.Pi / 180
	long1 := m1.Long * math.Pi / 180
	lat2 := m2.Lat * math.Pi / 180
	long2 := m2.Long * math.Pi / 180
	// obtained from https://en.wikipedia.org/wiki/Great-circle_distance
	deltaLat := lat2 - lat1
	deltaLong := long2 - long1
	centralAng := 2 * math.Asin(math.Sqrt(
		math.Pow(math.Sin(deltaLat/2), 2)+
			math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(deltaLong/2), 2)))
	return earthR * centralAng
}
