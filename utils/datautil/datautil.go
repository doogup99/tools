// Copyright © 2023 OpenIM. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datautil

import (
	"math/rand"
	"reflect"
	"sort"
	"time"

	"github.com/jinzhu/copier"

	"github.com/openimsdk/tools/db/pagination"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/utils/jsonutil"
)

// SliceSubFuncs returns elements in slice a that are not present in slice b (a - b) and remove duplicates.
// Determine if elements are equal based on the result returned by fna(a[i]) and fnb(b[i]).
func SliceSubFuncs[T, V any, E comparable](a []T, b []V, fna func(i T) E, fnb func(i V) E) []T {
	if len(b) == 0 {
		return a
	}
	k := make(map[E]struct{})
	for i := 0; i < len(b); i++ {
		k[fnb(b[i])] = struct{}{}
	}
	t := make(map[E]struct{})
	rs := make([]T, 0, len(a))
	for i := 0; i < len(a); i++ {
		e := fna(a[i])
		if _, ok := t[e]; ok {
			continue
		}
		if _, ok := k[e]; ok {
			continue
		}
		rs = append(rs, a[i])
		t[e] = struct{}{}
	}
	return rs
}

// SliceIntersectFuncs returns the intersection (a ∩ b) of slices a and b, removing duplicates.
// The equality of elements is determined by the custom functions fna and fnb provided for each slice.
func SliceIntersectFuncs[T, V any, E comparable](a []T, b []V, fna func(i T) E, fnb func(i V) E) []T {
	// If b is empty, return an empty slice
	if len(b) == 0 {
		return nil
	}

	// Map the elements of b to a set for fast lookup
	k := make(map[E]struct{})
	for _, item := range b {
		k[fnb(item)] = struct{}{}
	}

	// Create a slice to store the intersection result
	rs := make([]T, 0, len(a))
	// Map for deduplication
	t := make(map[E]struct{})

	// Iterate over slice a and find the common elements with b
	for _, item := range a {
		// Get the comparison value of the element from slice a
		e := fna(item)
		// If the element exists in both a and b, and hasn't been added yet, add it to the result
		if _, ok := k[e]; ok && t[e] == struct{}{} {
			rs = append(rs, item)
			t[e] = struct{}{}
		}
	}

	// Return the intersection slice
	return rs
}

// SliceSubFunc returns elements in slice a that are not present in slice b (a - b) and remove duplicates.
// Determine if elements are equal based on the result returned by fn.
func SliceSubFunc[T any, E comparable](a, b []T, fn func(i T) E) []T {
	return SliceSubFuncs(a, b, fn, fn)
}

// SliceSub returns elements in slice a that are not present in slice b (a - b) and remove duplicates.
func SliceSub[E comparable](a, b []E) []E {
	return SliceSubFunc(a, b, func(i E) E { return i })
}

// SliceSubAny returns elements in slice a that are not present in slice b (a - b) and remove duplicates.
// fn is a function that converts elements of slice b to elements comparable with those in slice a.
func SliceSubAny[E comparable, T any](a []E, b []T, fn func(t T) E) []E {
	return SliceSub(a, Slice(b, fn))
}

// SliceSubConvertPre returns elements in slice a that are not present in slice b (a - b) and remove duplicates.
// fn is a function that converts elements of slice a to elements comparable with those in slice b.
func SliceSubConvertPre[E comparable, T any](a []T, b []E, fn func(t T) E) []T {
	return SliceSubFuncs(a, b, fn, func(i E) E { return i })
}

// SliceAnySub returns elements in slice a that are not present in slice b (a - b).
func SliceAnySub[E any, T comparable](a, b []E, fn func(t E) T) []E {
	m := make(map[T]E)
	for i := 0; i < len(b); i++ {
		v := b[i]
		m[fn(v)] = v
	}
	var es []E
	for i := 0; i < len(a); i++ {
		v := a[i]
		if _, ok := m[fn(v)]; !ok {
			es = append(es, v)
		}
	}
	return es
}

// DistinctAny duplicate removal.
func DistinctAny[E any, K comparable](es []E, fn func(e E) K) []E {
	v := make([]E, 0, len(es))
	tmp := map[K]struct{}{}
	for i := 0; i < len(es); i++ {
		t := es[i]
		k := fn(t)
		if _, ok := tmp[k]; !ok {
			tmp[k] = struct{}{}
			v = append(v, t)
		}
	}
	return v
}

func DistinctAnyGetComparable[E any, K comparable](es []E, fn func(e E) K) []K {
	v := make([]K, 0, len(es))
	tmp := map[K]struct{}{}
	for i := 0; i < len(es); i++ {
		t := es[i]
		k := fn(t)
		if _, ok := tmp[k]; !ok {
			tmp[k] = struct{}{}
			v = append(v, k)
		}
	}
	return v
}

func Distinct[T comparable](ts []T) []T {
	if len(ts) < 2 {
		return ts
	} else if len(ts) == 2 {
		if ts[0] == ts[1] {
			return ts[:1]
		} else {
			return ts
		}
	}
	return DistinctAny(ts, func(t T) T {
		return t
	})
}

// Delete Delete slice elements, support negative number to delete the reciprocal number
func Delete[E any](es []E, index ...int) []E {
	switch len(index) {
	case 0:
		return es
	case 1:
		i := index[0]
		if i < 0 {
			i = len(es) + i
		}
		if len(es) <= i {
			return es
		}
		return append(es[:i], es[i+1:]...)
	default:
		tmp := make(map[int]struct{})
		for _, i := range index {
			if i < 0 {
				i = len(es) + i
			}
			tmp[i] = struct{}{}
		}
		v := make([]E, 0, len(es))
		for i := 0; i < len(es); i++ {
			if _, ok := tmp[i]; !ok {
				v = append(v, es[i])
			}
		}
		return v
	}
}

// DeleteAt Delete slice elements, support negative number to delete the reciprocal number
func DeleteAt[E any](es *[]E, index ...int) []E {
	v := Delete(*es, index...)
	*es = v
	return v
}

// IndexAny get the index of the element
func IndexAny[E any, K comparable](e E, es []E, fn func(e E) K) int {
	k := fn(e)
	for i := 0; i < len(es); i++ {
		if fn(es[i]) == k {
			return i
		}
	}
	return -1
}

// IndexOf get the index of the element
func IndexOf[E comparable](e E, es ...E) int {
	return IndexAny(e, es, func(t E) E {
		return t
	})
}

// DeleteElems delete elems in slice.
func DeleteElems[E comparable](es []E, delEs ...E) []E {
	switch len(delEs) {
	case 0:
		return es
	case 1:
		for i := range es {
			if es[i] == delEs[0] {
				return append(es[:i], es[i+1:]...)
			}
		}
		return es
	default:
		elMap := make(map[E]int)
		for _, e := range delEs {
			elMap[e]++
		}
		res := make([]E, 0, len(es))
		for i := range es {
			if _, ok := elMap[es[i]]; ok {
				elMap[es[i]]--
				if elMap[es[i]] == 0 {
					delete(elMap, es[i])
				}
				continue
			}
			res = append(res, es[i])
		}
		return res
	}
}

// Contain Whether to include
func Contain[E comparable](e E, es ...E) bool {
	return IndexOf(e, es...) >= 0
}

// Contains Whether to include
func Contains[E comparable](e []E, es ...E) bool {
	mp := SliceToMap(e, func(i E) E { return i })
	for _, e2 := range es {
		if _, ok := mp[e2]; ok {
			return true
		}
	}
	return false
}

// DuplicateAny Whether there are duplicates
func DuplicateAny[E any, K comparable](es []E, fn func(e E) K) bool {
	t := make(map[K]struct{})
	for _, e := range es {
		k := fn(e)
		if _, ok := t[k]; ok {
			return true
		}
		t[k] = struct{}{}
	}
	return false
}

// Duplicate Whether there are duplicates
func Duplicate[E comparable](es []E) bool {
	return DuplicateAny(es, func(e E) E {
		return e
	})
}

// SliceToMapOkAny slice to map (Custom type, filter)
func SliceToMapOkAny[E any, K comparable, V any](es []E, fn func(e E) (K, V, bool)) map[K]V {
	kv := make(map[K]V)
	for i := 0; i < len(es); i++ {
		t := es[i]
		if k, v, ok := fn(t); ok {
			kv[k] = v
		}
	}
	return kv
}

// SliceToMapAny slice to map (Custom type)
func SliceToMapAny[E any, K comparable, V any](es []E, fn func(e E) (K, V)) map[K]V {
	return SliceToMapOkAny(es, func(e E) (K, V, bool) {
		k, v := fn(e)
		return k, v, true
	})
}

// SliceToMap slice to map
func SliceToMap[E any, K comparable](es []E, fn func(e E) K) map[K]E {
	return SliceToMapOkAny(es, func(e E) (K, E, bool) {
		k := fn(e)
		return k, e, true
	})
}

// SliceSetAny slice to map[K]struct{}
func SliceSetAny[E any, K comparable](es []E, fn func(e E) K) map[K]struct{} {
	return SliceToMapAny(es, func(e E) (K, struct{}) {
		return fn(e), struct{}{}
	})
}

func Filter[E, T any](es []E, fn func(e E) (T, bool)) []T {
	rs := make([]T, 0, len(es))
	for i := 0; i < len(es); i++ {
		e := es[i]
		if t, ok := fn(e); ok {
			rs = append(rs, t)
		}
	}
	return rs
}

// Slice Converts slice types in batches
func Slice[E any, T any](es []E, fn func(e E) T) []T {
	v := make([]T, len(es))
	for i := 0; i < len(es); i++ {
		v[i] = fn(es[i])
	}
	return v
}

// SliceSet slice to map[E]struct{}
func SliceSet[E comparable](es []E) map[E]struct{} {
	return SliceSetAny(es, func(e E) E {
		return e
	})
}

// HasKey get whether the map contains key
func HasKey[K comparable, V any](m map[K]V, k K) bool {
	if m == nil {
		return false
	}
	_, ok := m[k]
	return ok
}

// Min get minimum value
func Min[E Ordered](e ...E) E {
	v := e[0]
	for _, t := range e[1:] {
		if v > t {
			v = t
		}
	}
	return v
}

// Max get maximum value
func Max[E Ordered](e ...E) E {
	v := e[0]
	for _, t := range e[1:] {
		if v < t {
			v = t
		}
	}
	return v
}

// Between checks if data is between left and right, excluding equality.
func Between[E Ordered](data, left, right E) bool {
	return left < data && data < right
}

// BetweenEq checks if data is between left and right, including equality.
func BetweenEq[E Ordered](data, left, right E) bool {
	return left <= data && data <= right
}

// BetweenLEq checks if data is between left and right, including left equality.
func BetweenLEq[E Ordered](data, left, right E) bool {
	return left <= data && data < right
}

// BetweenREq checks if data is between left and right, including right equality.
func BetweenREq[E Ordered](data, left, right E) bool {
	return left < data && data <= right
}

func Paginate[E any](es []E, pageNumber int, showNumber int) []E {
	if pageNumber <= 0 {
		return []E{}
	}
	if showNumber <= 0 {
		return []E{}
	}
	start := (pageNumber - 1) * showNumber
	end := start + showNumber
	if start >= len(es) {
		return []E{}
	}
	if end > len(es) {
		end = len(es)
	}
	return es[start:end]
}

func SlicePaginate[E any](es []E, pagination pagination.Pagination) []E {
	return Paginate(es, int(pagination.GetPageNumber()), int(pagination.GetShowNumber()))
}

// BothExistAny gets elements that are common in the slice (intersection)
func BothExistAny[E any, K comparable](es [][]E, fn func(e E) K) []E {
	if len(es) == 0 {
		return []E{}
	}
	var idx int
	ei := make([]map[K]E, len(es))
	for i := 0; i < len(ei); i++ {
		e := es[i]
		if len(e) == 0 {
			return []E{}
		}
		kv := make(map[K]E)
		for j := 0; j < len(e); j++ {
			t := e[j]
			k := fn(t)
			kv[k] = t
		}
		ei[i] = kv
		if len(kv) < len(ei[idx]) {
			idx = i
		}
	}
	v := make([]E, 0, len(ei[idx]))
	for k := range ei[idx] {
		all := true
		for i := 0; i < len(ei); i++ {
			if i == idx {
				continue
			}
			if _, ok := ei[i][k]; !ok {
				all = false
				break
			}
		}
		if !all {
			continue
		}
		v = append(v, ei[idx][k])
	}
	return v
}

// BothExist Gets the common elements in the slice (intersection)
func BothExist[E comparable](es ...[]E) []E {
	return BothExistAny(es, func(e E) E {
		return e
	})
}

//func CompleteAny[K comparable, E any](ks []K, es []E, fn func(e E) K) bool {
//	if len(ks) == 0 && len(es) == 0 {
//		return true
//	}
//	kn := make(map[K]uint8)
//	for _, e := range Distinct(ks) {
//		kn[e]++
//	}
//	for k := range SliceSetAny(es, fn) {
//		kn[k]++
//	}
//	for _, n := range kn {
//		if n != 2 {
//			return false
//		}
//	}
//	return true
//}

// Complete whether a and b are equal after deduplication (ignore order)
func Complete[E comparable](a []E, b []E) bool {
	return len(Single(a, b)) == 0
}

// Keys get map keys
func Keys[K comparable, V any](kv map[K]V) []K {
	ks := make([]K, 0, len(kv))
	for k := range kv {
		ks = append(ks, k)
	}
	return ks
}

// Values get map values
func Values[K comparable, V any](kv map[K]V) []V {
	vs := make([]V, 0, len(kv))
	for k := range kv {
		vs = append(vs, kv[k])
	}
	return vs
}

// Sort basic type sorting
func Sort[E Ordered](es []E, asc bool) []E {
	SortAny(es, func(a, b E) bool {
		if asc {
			return a < b
		} else {
			return a > b
		}
	})
	return es
}

// SortAny custom sort method
func SortAny[E any](es []E, fn func(a, b E) bool) {
	sort.Sort(&sortSlice[E]{
		ts: es,
		fn: fn,
	})
}

// If true -> a, false -> b
func If[T any](isa bool, a, b T) T {
	if isa {
		return a
	}
	return b
}

func ToPtr[T any](t T) *T {
	return &t
}

// Equal Compares slices to each other (including element order)
func Equal[E comparable](a []E, b []E) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Single exists in a and does not exist in b or exists in b and does not exist in a
func Single[E comparable](a, b []E) []E {
	kn := make(map[E]uint8)
	for _, e := range Distinct(a) {
		kn[e]++
	}
	for _, e := range Distinct(b) {
		kn[e]++
	}
	v := make([]E, 0, len(kn))
	for k, n := range kn {
		if n == 1 {
			v = append(v, k)
		}
	}
	return v
}

// Order sorts ts by es
func Order[E comparable, T any](es []E, ts []T, fn func(t T) E) []T {
	if len(es) == 0 || len(ts) == 0 {
		return ts
	}
	kv := make(map[E][]T)
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		k := fn(t)
		kv[k] = append(kv[k], t)
	}
	rs := make([]T, 0, len(ts))
	for _, e := range es {
		vs := kv[e]
		delete(kv, e)
		rs = append(rs, vs...)
	}
	for k := range kv {
		rs = append(rs, kv[k]...)
	}
	return rs
}

func OrderPtr[E comparable, T any](es []E, ts *[]T, fn func(t T) E) []T {
	*ts = Order(es, *ts, fn)
	return *ts
}

func UniqueJoin(s ...string) string {
	data, _ := jsonutil.JsonMarshal(s)
	return string(data)
}

type sortSlice[E any] struct {
	ts []E
	fn func(a, b E) bool
}

func (o *sortSlice[E]) Len() int {
	return len(o.ts)
}

func (o *sortSlice[E]) Less(i, j int) bool {
	return o.fn(o.ts[i], o.ts[j])
}

func (o *sortSlice[E]) Swap(i, j int) {
	o.ts[i], o.ts[j] = o.ts[j], o.ts[i]
}

// Ordered types that can be sorted
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

// NotNilReplace sets old to new_ when new_ is not null
func NotNilReplace[T any](old, new_ *T) {
	if new_ == nil {
		return
	}
	*old = *new_
}

func StructFieldNotNilReplace(dest, src any) {
	destVal := reflect.ValueOf(dest).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	for i := 0; i < destVal.NumField(); i++ {
		destField := destVal.Field(i)
		srcField := srcVal.Field(i)

		// Check if the source field is valid
		if srcField.IsValid() {
			// Check if the target field can be set
			if destField.CanSet() {
				// Handling fields of slice type
				if destField.Kind() == reflect.Slice && srcField.Kind() == reflect.Slice {
					elemType := destField.Type().Elem()
					// Check if a slice element is a pointer to a structure
					if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct {
						// Create a new slice to store the copied elements
						newSlice := reflect.MakeSlice(destField.Type(), srcField.Len(), srcField.Cap())
						for j := 0; j < srcField.Len(); j++ {
							newElem := reflect.New(elemType.Elem())
							// Recursive update, retaining non-zero values
							StructFieldNotNilReplace(newElem.Interface(), srcField.Index(j).Interface())
							// Checks if the field of the new element is zero-valued, and if so, preserves the value at the corresponding position in the original slice
							for k := 0; k < newElem.Elem().NumField(); k++ {
								if newElem.Elem().Field(k).IsZero() {
									newElem.Elem().Field(k).Set(destField.Index(j).Elem().Field(k))
								}
							}
							newSlice.Index(j).Set(newElem)
						}
						destField.Set(newSlice)
					} else {
						destField.Set(srcField)
					}
				} else {
					// For non-sliced fields, update the source field if it is non-zero, otherwise keep the original value
					if !srcField.IsZero() {
						destField.Set(srcField)
					}
				}
			}
		}
	}
}

func Batch[T any, V any](fn func(T) V, ts []T) []V {
	if ts == nil {
		return nil
	}
	res := make([]V, 0, len(ts))
	for i := range ts {
		res = append(res, fn(ts[i]))
	}
	return res
}

func InitSlice[T any](val *[]T) {
	if val != nil && *val == nil {
		*val = []T{}
	}
}

func InitMap[K comparable, V any](val *map[K]V) {
	if val != nil && *val == nil {
		*val = map[K]V{}
	}
}

func GetSwitchFromOptions(Options map[string]bool, key string) (result bool) {
	if Options == nil {
		return true
	}
	if flag, ok := Options[key]; !ok || flag {
		return true
	}
	return false
}

func SetSwitchFromOptions(options map[string]bool, key string, value bool) {
	if options == nil {
		options = make(map[string]bool, 5)
	}
	options[key] = value
}

// copy a by b  b->a
func CopyStructFields(a any, b any, fields ...string) (err error) {
	return copier.Copy(a, b)
}

func CopySlice[T any](a []T) []T {
	ns := make([]T, len(a))
	copy(ns, a)
	return ns
}

func ShuffleSlice[T any](a []T) []T {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shuffled := CopySlice(a)
	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func GetElemByIndex(array []int, index int) (int, error) {
	if index < 0 || index >= len(array) {
		return 0, errs.New("index out of range", "index", index, "array", array).Wrap()
	}

	return array[index], nil
}
