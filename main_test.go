/* Copyright 2021 Ben Fuller
 * Apache License, Version 2.0
 * See LICENCE file for copyright and licence details.
 */

package main

import "testing"

func TestDmsToDec(t *testing.T) {
	dms := "50|30"
	var exp float64 = 50.5
	got := DmsToDec(dms)
	if got != exp {
		t.Errorf("DmsToDec(%s) = %f; wanted %f", dms, got, exp)
	}
}

func TestFindRanges(t *testing.T) {
	m := Markers{
		Markers: []Marker{{
			Name: "a",
			Lat:  10,
			Long: -10,
		}, {
			Name: "b",
			Lat:  20,
			Long: -15,
		}, {
			Name: "c",
			Lat:  15,
			Long: -20,
		}},
	}

	latMax := m.FindRanges(true, true)
	latMin := m.FindRanges(true, false)
	longMax := m.FindRanges(false, true)
	longMin := m.FindRanges(false, false)

	if latMax != 1 {
		t.Errorf("latMax = %v, wanted 1", latMax)
	}
	if latMin != 0 {
		t.Errorf("latMin = %v, wanted 0", latMin)
	}
	if longMax != 0 {
		t.Errorf("longMax = %v, wanted 0", longMax)
	}
	if longMin != 2 {
		t.Errorf("longMin = %v, wanted 2", longMin)
	}
}
