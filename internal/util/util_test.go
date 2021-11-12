package util_test

import (
	"activity_log/internal/util"
	"testing"
)

func TestNestedMapsEqual(t *testing.T) {
	map1 := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": map[string]interface{}{
				"Philip": nil,
				"Mary":   nil,
			},
			"JohnAndAnn": map[string]interface{}{
				"Luke":  nil,
				"Mark":  nil,
				"Simon": 3,
			},
		},
	}

	map1Copy, err := util.CopyNestedMap(map1)
	if err != nil {
		t.Fatalf("CopyNestedMap() returns err: %v", err)
	}

	if err := util.NestedMapsEqual(map1, map1Copy); err != nil {
		t.Fatalf("NestedMapsEqual() returns err for equal maps: %v", err)
	}

	if err := util.NestedMapsEqual(map1, map[string]interface{}{}); err == nil {
		t.Fatalf("NestedMapsEqual() returns equal for full map vs empty map")
	}

	map1Copy["SomeoneElse"] = map[string]interface{}{"new": 5}
	if err := util.NestedMapsEqual(map1, map1Copy); err == nil {
		t.Fatalf("NestedMapsEqual() returns equal for two different maps")
	}
}

func TestNewExpandingMap(t *testing.T) {
	goodMap := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": map[string]interface{}{
				"Philip": nil,
				"Mary":   nil,
			},
			"JohnAndAnn": map[string]interface{}{
				"Luke":  nil,
				"Mark":  nil,
				"Simon": nil,
			},
		},
	}

	expandingMap, err := util.NewExpandingMap(goodMap)
	if err != nil {
		t.Fatalf("util.NewExpandingMap() returns err: %v", err)
	}

	if err := util.NestedMapsEqual(goodMap, expandingMap.ToRegularMap()); err != nil {
		t.Errorf("Maps differ: want: %+v, got: %+v, err: %v", goodMap, expandingMap.ToRegularMap(), err)
	}

	badMap := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": 5,
		},
	}
	_, err = util.NewExpandingMap(badMap)
	if err == nil {
		t.Fatalf("util.NewExpandingMap() should've returned err")
	}

}

func TestExpandingMapsEqual(t *testing.T) {

	mapPreLoad1 := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": map[string]interface{}{
				"Philip": nil,
				"Mary":   nil,
			},
			"JohnAndAnn": map[string]interface{}{
				"Luke":  nil,
				"Mark":  nil,
				"Simon": nil,
			},
		},
	}

	map1, err := util.NewExpandingMap(mapPreLoad1)
	if err != nil {
		t.Fatalf("NewExpandingMap(mapPreLoad1) returns err: %v", err)
	}
	map1Clone, err := util.NewExpandingMap(mapPreLoad1)
	if err != nil {
		t.Fatalf("NewExpandingMap(mapPreLoad1) returns err: %v", err)
	}

	mapPreLoad1["SomeoneElse"] = map[string]interface{}{"other": nil}
	mapDifferent, err := util.NewExpandingMap(mapPreLoad1)
	if err != nil {
		t.Fatalf("NewExpandingMap(mapPreLoad1) returns err: %v", err)
	}

	mapEmpty, err := util.NewExpandingMap(nil)
	if err != nil {
		t.Fatalf("NewExpandingMap(mapPreLoad1) returns err: %v", err)
	}

	testCases := []struct {
		desc            string
		lhsExpandingMap *util.ExpandingMap
		rhsExpandingMap *util.ExpandingMap
		areEqual        bool
	}{
		{
			desc:            "large equal maps",
			lhsExpandingMap: map1,
			rhsExpandingMap: map1Clone,
			areEqual:        true,
		},
		{
			desc:            "large different maps",
			lhsExpandingMap: map1,
			rhsExpandingMap: mapDifferent,
			areEqual:        false,
		},
		{
			desc:            "large map vs empty map",
			lhsExpandingMap: map1,
			rhsExpandingMap: mapEmpty,
			areEqual:        false,
		},
		{
			desc:            "large map vs nil",
			lhsExpandingMap: map1,
			rhsExpandingMap: nil,
			areEqual:        false,
		},
		{
			desc:            "two empty maps",
			lhsExpandingMap: mapEmpty,
			rhsExpandingMap: mapEmpty,
			areEqual:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.lhsExpandingMap.IsEqual(tc.rhsExpandingMap)

			if tc.areEqual && err != nil {
				t.Errorf("not equal: %v", err)
			}

			if !tc.areEqual && err == nil {
				t.Errorf("should've been considered not equal")
			}
		})
	}

}

func TestAddSubMap(t *testing.T) {
	mapBeforeDict := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": map[string]interface{}{
				"Philip": nil,
				"Mary":   nil,
			},
			"JohnAndAnn": map[string]interface{}{
				"Luke":  nil,
				"Mark":  nil,
				"Simon": nil,
			},
		},
	}

	mapBefore, err := util.NewExpandingMap(mapBeforeDict)
	if err != nil {
		t.Fatalf("NewExpandingMap(mapPreLoad1) returns err: %v", err)
	}

	if err := mapBefore.AddSubMap([]string{"RyanAndViolet", "JohnAndAnn", "Luke"}, "Polo"); err != nil {
		t.Fatalf("AddSubMap() returns err: %v", err)
	}

	mapAfterDict := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": map[string]interface{}{
				"Philip": nil,
				"Mary":   nil,
			},
			"JohnAndAnn": map[string]interface{}{
				"Luke": map[string]interface{}{
					"Polo": nil,
				},
				"Mark":  nil,
				"Simon": nil,
			},
		},
	}

	mapAfter, err := util.NewExpandingMap(mapAfterDict)
	if err != nil {
		t.Fatalf("NewExpandingMap(mapPreLoad1) returns err: %v", err)
	}

	if err := mapAfter.IsEqual(mapBefore); err != nil {
		t.Fatalf("Not equal: %v", err)
	}

	gotSubMap, err := mapBefore.GetSubMap([]string{"RyanAndViolet", "JohnAndAnn", "Luke", "Polo"})
	if err != nil {
		t.Fatalf("GetSubMap() returns err: %v", err)
	}

	if !gotSubMap.IsEmpty() {
		t.Fatalf("submap should be empty")
	}
}

func TestGetSubMap(t *testing.T) {
	mapDict := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": map[string]interface{}{
				"Philip": nil,
				"Mary":   nil,
			},
			"JohnAndAnn": map[string]interface{}{
				"Luke": map[string]interface{}{
					"Polo": nil,
				},
				"Mark":  nil,
				"Simon": nil,
			},
		},
	}

	subMapDict := map[string]interface{}{
		"Polo": nil,
	}

	eMap, err := util.NewExpandingMap(mapDict)
	if err != nil {
		t.Fatalf("NewExpandingMap(mapDict) returns err: %v", err)
	}

	expectedSubMap, err := util.NewExpandingMap(subMapDict)
	if err != nil {
		t.Fatalf("NewExpandingMap(subMapDict) returns err: %v", err)
	}

	subMap, err := eMap.GetSubMap([]string{"RyanAndViolet", "JohnAndAnn", "Luke"})
	if err != nil {
		t.Fatalf("GetSubMap() returns err: %v", err)
	}

	if err := expectedSubMap.IsEqual(subMap); err != nil {
		t.Fatalf("Maps not equal: %v", err)
	}
}

func TestConstructionFromScratch(t *testing.T) {
	mapDict := map[string]interface{}{
		"SomeoneElse": nil,
		"RyanAndViolet": map[string]interface{}{
			"SimonAndTanya": map[string]interface{}{
				"Philip": nil,
				"Mary":   nil,
			},
			"JohnAndAnn": map[string]interface{}{
				"Luke": map[string]interface{}{
					"Luke": nil,
					"Polo": nil,
				},
				"Mark":  nil,
				"Simon": nil,
			},
		},
	}

	expectedEMap, err := util.NewExpandingMap(mapDict)
	if err != nil {
		t.Fatalf("NewExpandingMap() returns err: %v", err)
	}

	eMap := util.NewEmptyExpandingMap()

	if err := eMap.AddSubMap([]string{}, "SomeoneElse"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{}, "SomeoneElse", err)
	}
	if err := eMap.AddSubMap([]string{}, "RyanAndViolet"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{}, "RyanAndViolet", err)
	}
	if err := eMap.AddSubMap([]string{"RyanAndViolet"}, "SimonAndTanya"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet"}, "SimonAndTanya", err)
	}
	if err := eMap.AddSubMap([]string{"RyanAndViolet"}, "JohnAndAnn"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet"}, "JohnAndAnn", err)
	}
	if err := eMap.AddSubMap([]string{"RyanAndViolet", "SimonAndTanya"}, "Philip"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet", "SimonAndTanya"}, "Philip", err)
	}
	if err := eMap.AddSubMap([]string{"RyanAndViolet", "SimonAndTanya"}, "Mary"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet", "SimonAndTanya"}, "Philip", err)
	}
	if err := eMap.AddSubMap([]string{"RyanAndViolet", "JohnAndAnn"}, "Luke"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet", "JohnAndAnn"}, "Luke", err)
	}
	if err := eMap.AddSubMapIncludingParent([]string{"RyanAndViolet", "JohnAndAnn", "Luke"}, "Polo"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet", "JohnAndAnn", "Luke"}, "Polo", err)
	}
	if err := eMap.AddSubMap([]string{"RyanAndViolet", "JohnAndAnn"}, "Mark"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet", "JohnAndAnn"}, "Mark", err)
	}
	if err := eMap.AddSubMap([]string{"RyanAndViolet", "JohnAndAnn"}, "Simon"); err != nil {
		t.Fatalf("AddSubMap(%v, %s) returns err: %v", []string{"RyanAndViolet", "JohnAndAnn"}, "Simon", err)
	}

	if err := expectedEMap.IsEqual(eMap); err != nil {
		t.Fatalf("maps not equal: %v", err)
	}
}
