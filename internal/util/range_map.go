package util

import (
	"fmt"
	"seneca/api/senecaerror"
	"sort"
)

// [L, U).
type Range struct {
	L int64
	U int64
}

func (r Range) String() string {
	return fmt.Sprintf("[%v, %v)", r.L, r.U)
}

// TODO(lucaloncar): enforce invariants (such as ordered non-overlapping keys)  with a New() function.
type RangeMap struct {
	keys   []Range
	values []interface{}
}

func NewRangeMap(keys []Range, values []interface{}) (*RangeMap, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("keys length %d not equal to values length %d", len(keys), len(values))
	}

	if keys == nil {
		keys = []Range{}
		values = []interface{}{}
	}

	rm := &RangeMap{
		keys:   keys,
		values: values,
	}

	for i := range keys {
		if err := rm.Insert(keys[i], values[i]); err != nil {
			return nil, fmt.Errorf("NewRangeMap().Insert(%s, _) returns err: %w", keys[i], err)
		}
	}

	return rm, nil
}

func (rm *RangeMap) Length() int {
	return len(rm.keys)
}

func (rm *RangeMap) String() string {
	output := "{ "
	for i := range rm.keys {
		output += fmt.Sprintf("  (%s, _),  ", rm.keys[i])
	}
	output += " }"
	return output
}

func (rm *RangeMap) Get(key int64) (interface{}, bool) {
	if rm == nil {
		return nil, false
	}

	if rm.keys == nil {
		return nil, false
	}

	i := sort.Search(len(rm.keys), func(i int) bool {
		return key < rm.keys[i].L
	})

	i--
	if i >= 0 && i < len(rm.keys) && key < rm.keys[i].U {
		return rm.values[i], true
	}
	return nil, false
}

func (rm *RangeMap) Insert(keyRange Range, obj interface{}) error {
	if rm == nil || rm.keys == nil || rm.values == nil {
		return senecaerror.NewDevError(fmt.Errorf("RangeMap not initialized properly.  Call NewRangeMap()"))
	}

	i := sort.Search(len(rm.keys), func(i int) bool {
		return keyRange.L < rm.keys[i].L
	})

	i--
	// Value was found.
	if i >= 0 && i < len(rm.keys) {
		// Upper bound overlap.
		if i < len(rm.keys)-1 && keyRange.U > rm.keys[i+1].L {
			return fmt.Errorf("new range %q upper bound overlaps with existing range %q lower bound", keyRange, rm.keys[i+1])
		}

		// Lower bound overlap.
		if i > 0 && keyRange.L < rm.keys[i-1].U {
			return fmt.Errorf("new range %q lower bound overlaps with existing range %q upper bound", keyRange, rm.keys[i-1])
		}

		// Replace.
		if keyRange.L == rm.keys[i].L && keyRange.U == rm.keys[i].U {
			rm.values[i] = obj
			return nil
		}
	}

	i++
	rm.keys = append(rm.keys, keyRange)
	copy(rm.keys[(i+1):], rm.keys[i:])
	rm.keys[i] = keyRange

	rm.values = append(rm.values, obj)
	copy(rm.values[(i+1):], rm.values[i:])
	rm.values[i] = obj

	return nil
}
