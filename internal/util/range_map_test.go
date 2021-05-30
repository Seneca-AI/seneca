package util

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestRangeMap(t *testing.T) {
	keys := []Range{
		{
			L: -30,
			U: -20,
		},
		{
			L: -20,
			U: -10,
		},
		{
			L: -10,
			U: 0,
		},
		{
			L: 0,
			U: 10,
		},
		{
			L: 20,
			U: 30,
		},
	}
	values := []interface{}{"one", "two", "three", "four", "six"}

	rangeMap, err := NewRangeMap(keys, values)
	if err != nil {
		t.Fatalf("NewRangeMap() returns err: %v", err)
	}

	valObj, ok := rangeMap.Get(5)
	if !ok {
		t.Fatalf("5 not found in RangeMap")
	}
	val, ok := valObj.(string)
	if !ok {
		t.Fatalf("Want string, got %T", valObj)
	}
	if val != "four" {
		t.Fatalf("Want %q, got %q. RangeMap: %v", "four", val, rangeMap)
	}

	if err := rangeMap.Insert(Range{L: 10, U: 20}, "five"); err != nil {
		t.Fatalf("rangeMap.Insert() returns err: %v", err)
	}

	valObj, ok = rangeMap.Get(15)
	if !ok {
		t.Fatalf("15 not found in RangeMap. RangeMap: %v", rangeMap)
	}
	val, ok = valObj.(string)
	if !ok {
		t.Fatalf("Want string, got %T", valObj)
	}
	if val != "five" {
		t.Fatalf("Want %q, got %q", "five", val)
	}

	rangeMap.Insert(Range{L: 30, U: 40}, "seven")

	valObj, ok = rangeMap.Get(35)
	if !ok {
		t.Fatalf("35 not found in RangeMap. RangeMap: %v", rangeMap)
	}
	val, ok = valObj.(string)
	if !ok {
		t.Fatalf("Want string, got %T", valObj)
	}
	if val != "seven" {
		t.Fatalf("Want %q, got %q", "seven", val)
	}

}

func TestRangeMapExhaustive(t *testing.T) {
	rangeAndValues := []rangeAndValue{}
	for i := -200; i < 200; i += 10 {
		rnv := rangeAndValue{
			r: Range{L: int64(i), U: int64(i) + 10},
			v: fmt.Sprintf("%d", i),
		}
		rangeAndValues = append(rangeAndValues, rnv)
	}

	for j := 0; j < 10; j++ {
		rand.Seed(time.Now().UnixNano() + int64(j))
		rand.Shuffle(len(rangeAndValues), func(i, j int) { rangeAndValues[i], rangeAndValues[j] = rangeAndValues[j], rangeAndValues[i] })

		rangeMap, err := NewRangeMap(nil, nil)
		if err != nil {
			t.Fatalf("NewRangeMap() returns err: %v", err)
		}

		for _, rnv := range rangeAndValues {
			if err := rangeMap.Insert(rnv.r, rnv.v); err != nil {
				t.Fatalf("Insert(%s, %s) returns err: %v", rnv.r, rnv.v, err)
			}
		}

		rand.Seed(time.Now().UnixNano() + int64(j))
		rand.Shuffle(len(rangeAndValues), func(i, j int) { rangeAndValues[i], rangeAndValues[j] = rangeAndValues[j], rangeAndValues[i] })
		for _, rnv := range rangeAndValues {
			if err := rangeMap.Insert(rnv.r, rnv.v); err != nil {
				t.Fatalf("Insert(%s, %s) returns err: %v", rnv.r, rnv.v, err)
			}
		}

		if rangeMap.Length() != len(rangeAndValues) {
			t.Fatalf("Want length %d for rangeMap, got len %d", len(rangeAndValues), rangeMap.Length())
		}

		for i := -200; i < 200; i++ {
			out, ok := rangeMap.Get(int64(i))
			if !ok {
				t.Fatalf("Could not find %d in rangeMap %v", i, rangeMap)
			}

			bottomOfRange := i
			modulo := i % 10
			if modulo < 0 {
				bottomOfRange -= 10
			}
			bottomOfRange = bottomOfRange - modulo

			if out != fmt.Sprintf("%d", bottomOfRange) {
				t.Fatalf("Want %q for %d, got %q in rangeMap %v", fmt.Sprintf("%d", bottomOfRange), i, out, rangeMap)
			}
		}
	}
}

type rangeAndValue struct {
	r Range
	v string
}
