// Copyright 2021 dudaodong@gmail.com. All rights reserved.
// Use of this source code is governed by MIT license

// Package slice implements some functions to manipulate slice.
package slice

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"unsafe"
)

// Contain check if the value is in the iterable type or not
func Contain[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

// Chunk creates an slice of elements split into groups the length of size.
func Chunk[T any](slice []T, size int) [][]T {
	var res [][]T

	if len(slice) == 0 || size <= 0 {
		return res
	}

	length := len(slice)
	if size == 1 || size >= length {
		for _, v := range slice {
			var tmp []T
			tmp = append(tmp, v)
			res = append(res, tmp)
		}
		return res
	}

	// divide slice equally
	divideNum := length/size + 1
	for i := 0; i < divideNum; i++ {
		if i == divideNum-1 {
			if len(slice[i*size:]) > 0 {
				res = append(res, slice[i*size:])
			}
		} else {
			res = append(res, slice[i*size:(i+1)*size])
		}
	}

	return res
}

// Difference creates an slice of whose element in slice1 but not in slice2
func Difference[T comparable](slice1, slice2 []T) []T {
	var res []T
	for _, v := range slice1 {
		if !Contain(slice2, v) {
			res = append(res, v)
		}
	}

	return res
}

// Every return true if all of the values in the slice pass the predicate function.
// The fn function signature should be func(int, T) bool .
func Every[T any](slice []T, fn func(index int, t T) bool) bool {
	var currentLength int

	for i, v := range slice {
		if fn(i, v) {
			currentLength++
		}
	}

	return currentLength == len(slice)
}

// None return true if all the values in the slice mismatch the criteria
// The fn function signature should be func(int, T) bool .
func None[T any](slice []T, fn func(index int, t T) bool) bool {
	var currentLength int

	for i, v := range slice {
		if !fn(i, v) {
			currentLength++
		}
	}

	return currentLength == len(slice)
}

// Some return true if any of the values in the list pass the predicate function.
// The fn function signature should be func(int, T) bool.
func Some[T any](slice []T, fn func(index int, t T) bool) bool {
	for i, v := range slice {
		if fn(i, v) {
			return true
		}
	}
	return false
}

// Filter iterates over elements of slice, returning an slice of all elements `signature` returns truthy for.
// The fn function signature should be func(int, T) bool.
func Filter[T any](slice []T, fn func(index int, t T) bool) []T {
	res := make([]T, 0, 0)
	for i, v := range slice {
		if fn(i, v) {
			res = append(res, v)
		}
	}
	return res
}

// Count iterates over elements of slice, returns a count of all matched elements
// The function signature should be func(index int, value interface{}) bool .
func Count[T any](slice []T, fn func(index int, t T) bool) int {
	if fn == nil {
		panic("fn is missing")
	}

	if len(slice) == 0 {
		return 0
	}

	var count int
	for i, v := range slice {
		if fn(i, v) {
			count++
		}
	}

	return count
}

// GroupBy iterate over elements of the slice, each element will be group by criteria, returns two slices
// The function signature should be func(index int, value interface{}) bool .
func GroupBy(slice, function interface{}) (interface{}, interface{}) {
	sv := sliceValue(slice)
	fn := functionValue(function)

	elemType := sv.Type().Elem()
	if checkSliceCallbackFuncSignature(fn, elemType, reflect.ValueOf(true).Type()) {
		panic("function param should be of type func(int, " + elemType.String() + ")" + reflect.ValueOf(true).Type().String())
	}

	groupB := reflect.MakeSlice(sv.Type(), 0, 0)
	groupA := reflect.MakeSlice(sv.Type(), 0, 0)

	for i := 0; i < sv.Len(); i++ {
		flag := fn.Call([]reflect.Value{reflect.ValueOf(i), sv.Index(i)})[0]
		if flag.Bool() {
			groupA = reflect.Append(groupA, sv.Index(i))
		} else {
			groupB = reflect.Append(groupB, sv.Index(i))
		}
	}

	return groupA.Interface(), groupB.Interface()
}

// Find iterates over elements of slice, returning the first one that passes a truth test on function.
// The fn function signature should be func(int, T) bool .
// If return T is nil then no items matched the predicate func
func Find[T any](slice []T, fn func(index int, t T) bool) (*T, bool) {
	if len(slice) == 0 {
		return nil, false
	}

	index := -1
	for i, v := range slice {
		if fn(i, v) {
			index = i
			break
		}
	}

	if index == -1 {
		return nil, false
	}

	return &slice[index], true
}

// FlattenDeep flattens slice recursive
func FlattenDeep(slice interface{}) interface{} {
	sv := sliceValue(slice)
	st := sliceElemType(sv.Type())
	tmp := reflect.MakeSlice(reflect.SliceOf(st), 0, 0)
	res := flattenRecursive(sv, tmp)
	return res.Interface()
}

func flattenRecursive(value reflect.Value, result reflect.Value) reflect.Value {
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		kind := item.Kind()

		if kind == reflect.Slice {
			result = flattenRecursive(item, result)
		} else {
			result = reflect.Append(result, item)
		}
	}

	return result
}

// ForEach iterates over elements of slice and invokes function for each element
// The fn signature should be func(int, T).
func ForEach[T any](slice []T, fn func(index int, t T)) {
	for i, v := range slice {
		fn(i, v)
	}
}

// Map creates an slice of values by running each element of slice thru fn function.
// The fn signature should be func(int, T).
func Map[T any, U any](slice []T, fn func(index int, t T) U) []U {
	res := make([]U, len(slice), cap(slice))
	for i, v := range slice {
		res[i] = fn(i, v)
	}

	return res
}

// Reduce creates an slice of values by running each element of slice thru fn function.
// The fn function signature should be fn func(int, T) T .
func Reduce[T any](slice []T, fn func(index int, t1, t2 T) T, initial T) T {
	if len(slice) == 0 {
		return initial
	}

	res := fn(0, initial, slice[0])

	tmp := make([]T, 2, 2)
	for i := 1; i < len(slice); i++ {
		tmp[0] = res
		tmp[1] = slice[i]
		res = fn(i, tmp[0], tmp[1])
	}

	return res
}

// InterfaceSlice convert param to slice of interface.
func InterfaceSlice(slice interface{}) []interface{} {
	sv := sliceValue(slice)
	if sv.IsNil() {
		return nil
	}

	res := make([]interface{}, sv.Len())
	for i := 0; i < sv.Len(); i++ {
		res[i] = sv.Index(i).Interface()
	}

	return res
}

// StringSlice convert param to slice of string.
func StringSlice(slice interface{}) []string {
	v := sliceValue(slice)

	out := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		v, ok := v.Index(i).Interface().(string)
		if !ok {
			panic("invalid element type")
		}
		out[i] = v
	}

	return out
}

// IntSlice convert param to slice of int.
func IntSlice(slice interface{}) []int {
	sv := sliceValue(slice)

	out := make([]int, sv.Len())
	for i := 0; i < sv.Len(); i++ {
		v, ok := sv.Index(i).Interface().(int)
		if !ok {
			panic("invalid element type")
		}
		out[i] = v
	}

	return out
}

// ConvertSlice convert original slice to new data type element of slice.
func ConvertSlice(originalSlice interface{}, newSliceType reflect.Type) interface{} {
	sv := sliceValue(originalSlice)
	if newSliceType.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Invalid newSliceType(non-slice type of type %T)", newSliceType))
	}

	newSlice := reflect.New(newSliceType)

	hdr := (*reflect.SliceHeader)(unsafe.Pointer(newSlice.Pointer()))

	var newElemSize = int(sv.Type().Elem().Size()) / int(newSliceType.Elem().Size())

	hdr.Cap = sv.Cap() * newElemSize
	hdr.Len = sv.Len() * newElemSize
	hdr.Data = sv.Pointer()

	return newSlice.Elem().Interface()
}

// DeleteByIndex delete the element of slice from start index to end index - 1.
// Delete i: s = append(s[:i], s[i+1:]...)
// Delete i to j: s = append(s[:i], s[j:]...)
func DeleteByIndex(slice interface{}, start int, end ...int) (interface{}, error) {
	v := sliceValue(slice)
	i := start
	if v.Len() == 0 || i < 0 || i > v.Len() {
		return nil, errors.New("InvalidStartIndex")
	}
	if len(end) > 0 {
		j := end[0]
		if j <= i || j > v.Len() {
			return nil, errors.New("InvalidEndIndex")
		}
		v = reflect.AppendSlice(v.Slice(0, i), v.Slice(j, v.Len()))
	} else {
		v = reflect.AppendSlice(v.Slice(0, i), v.Slice(i+1, v.Len()))
	}

	return v.Interface(), nil
}

// Drop creates a slice with `n` elements dropped from the beginning when n > 0, or `n` elements dropped from the ending when n < 0
func Drop(slice interface{}, n int) interface{} {
	sv := sliceValue(slice)

	if n == 0 {
		return slice
	}

	svLen := sv.Len()

	if math.Abs(float64(n)) >= float64(svLen) {
		return reflect.MakeSlice(sv.Type(), 0, 0).Interface()
	}

	if n > 0 {
		res := reflect.MakeSlice(sv.Type(), svLen-n, svLen-n)
		for i := 0; i < res.Len(); i++ {
			res.Index(i).Set(sv.Index(i + n))
		}

		return res.Interface()
	}

	res := reflect.MakeSlice(sv.Type(), svLen+n, svLen+n)
	for i := 0; i < res.Len(); i++ {
		res.Index(i).Set(sv.Index(i))
	}

	return res.Interface()
}

// InsertByIndex insert the element into slice at index.
// Insert value: s = append(s[:i], append([]T{x}, s[i:]...)...)
// Insert slice: a = append(a[:i], append(b, a[i:]...)...)
func InsertByIndex(slice interface{}, index int, value interface{}) (interface{}, error) {
	v := sliceValue(slice)

	if index < 0 || index > v.Len() {
		return slice, errors.New("InvalidSliceIndex")
	}

	// value is slice
	vv := reflect.ValueOf(value)
	if vv.Kind() == reflect.Slice {
		if reflect.TypeOf(slice).Elem() != reflect.TypeOf(value).Elem() {
			return slice, errors.New("InvalidValueType")
		}
		v = reflect.AppendSlice(v.Slice(0, index), reflect.AppendSlice(vv.Slice(0, vv.Len()), v.Slice(index, v.Len())))
		return v.Interface(), nil
	}

	// value is not slice
	if reflect.TypeOf(slice).Elem() != reflect.TypeOf(value) {
		return slice, errors.New("InvalidValueType")
	}
	if index == v.Len() {
		return reflect.Append(v, reflect.ValueOf(value)).Interface(), nil
	}

	v = reflect.AppendSlice(v.Slice(0, index+1), v.Slice(index, v.Len()))
	v.Index(index).Set(reflect.ValueOf(value))

	return v.Interface(), nil
}

// UpdateByIndex update the slice element at index.
func UpdateByIndex(slice interface{}, index int, value interface{}) (interface{}, error) {
	v := sliceValue(slice)

	if index < 0 || index >= v.Len() {
		return slice, errors.New("InvalidSliceIndex")
	}

	if reflect.TypeOf(slice).Elem() != reflect.TypeOf(value) {
		return slice, errors.New("InvalidValueType")
	}

	v.Index(index).Set(reflect.ValueOf(value))

	return v.Interface(), nil
}

// Unique remove duplicate elements in slice.
func Unique[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return []T{}
	}

	// here no use map filter. if use it, the result slice element order is random, not same as origin slice
	var res []T
	for i := 0; i < len(slice); i++ {
		v := slice[i]
		skip := true
		for j := range res {
			if v == res[j] {
				skip = false
				break
			}
		}
		if skip {
			res = append(res, v)
		}
	}

	return res
}

// Union creates a slice of unique values, in order, from all given slices. using == for equality comparisons.
func Union[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}

	// append all slices, then unique it
	var allElements []T

	for _, slice := range slices {
		for _, v := range slice {
			allElements = append(allElements, v)
		}
	}

	return Unique(allElements)
}

// Intersection creates a slice of unique values that included by all slices.
func Intersection[T comparable](slices ...[]T) []T {
	var res []T
	if len(slices) == 0 {
		return []T{}
	}
	if len(slices) == 1 {
		return Unique(slices[0])
	}

	//return elements both in slice1 and slice2
	reduceFunc := func(slice1, slice2 []T) []T {
		s := make([]T, 0, 0)
		for _, v := range slice1 {
			if Contain(slice2, v) {
				s = append(s, v)
			}
		}
		return s
	}

	res = reduceFunc(slices[0], slices[1])

	if len(slices) == 2 {
		return Unique(res)
	}

	tmp := make([][]T, 2, 2)
	for i := 2; i < len(slices); i++ {
		tmp[0] = res
		tmp[1] = slices[i]
		res = reduceFunc(tmp[0], tmp[1])
	}

	return Unique(res)
}

// ReverseSlice return slice of element order is reversed to the given slice
func ReverseSlice(slice interface{}) {
	sv := sliceValue(slice)
	swp := reflect.Swapper(sv.Interface())
	for i, j := 0, sv.Len()-1; i < j; i, j = i+1, j-1 {
		swp(i, j)
	}
}

// Shuffle creates an slice of shuffled values
func Shuffle(slice interface{}) interface{} {
	sv := sliceValue(slice)
	length := sv.Len()

	res := reflect.MakeSlice(sv.Type(), length, length)
	for i, v := range rand.Perm(length) {
		res.Index(i).Set(sv.Index(v))
	}

	return res.Interface()
}

// SortByField return sorted slice by field
// Slice element should be struct, field type should be int, uint, string, or bool
// default sortType is ascending (asc), if descending order, set sortType to desc
func SortByField(slice interface{}, field string, sortType ...string) error {
	sv := sliceValue(slice)
	t := sv.Type().Elem()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("data type %T not support, shuld be struct or pointer to struct", slice)
	}

	// Find the field.
	sf, ok := t.FieldByName(field)
	if !ok {
		return fmt.Errorf("field name %s not found", field)
	}

	// Create a less function based on the field's kind.
	var less func(a, b reflect.Value) bool
	switch sf.Type.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		less = func(a, b reflect.Value) bool { return a.Int() < b.Int() }
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		less = func(a, b reflect.Value) bool { return a.Uint() < b.Uint() }
	case reflect.Float32, reflect.Float64:
		less = func(a, b reflect.Value) bool { return a.Float() < b.Float() }
	case reflect.String:
		less = func(a, b reflect.Value) bool { return a.String() < b.String() }
	case reflect.Bool:
		less = func(a, b reflect.Value) bool { return !a.Bool() && b.Bool() }
	default:
		return fmt.Errorf("field type %s not supported", sf.Type)
	}

	sort.Slice(slice, func(i, j int) bool {
		a := sv.Index(i)
		b := sv.Index(j)
		if t.Kind() == reflect.Ptr {
			a = a.Elem()
			b = b.Elem()
		}
		a = a.FieldByIndex(sf.Index)
		b = b.FieldByIndex(sf.Index)
		return less(a, b)
	})

	if sortType[0] == "desc" {
		ReverseSlice(slice)
	}
	return nil
}

// Without creates a slice excluding all given values
func Without(slice interface{}, values ...interface{}) interface{} {
	sv := sliceValue(slice)
	if sv.Len() == 0 {
		return slice
	}

	var indexes []int
	for i := 0; i < sv.Len(); i++ {
		v := sv.Index(i).Interface()
		if !Contain(values, v) {
			indexes = append(indexes, i)
		}
	}

	res := reflect.MakeSlice(sv.Type(), len(indexes), len(indexes))
	for i := range indexes {
		res.Index(i).Set(sv.Index(indexes[i]))
	}

	return res.Interface()
}
