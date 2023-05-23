package errors

import (
	"reflect"
	"sort"
)

// Empty 是公共的，因为它被一些内部API对象用于在外部字符串数组和内部集之间进行转换，并且转换逻辑现在需要公共类型.
type Empty struct{}

// String 是一组字符串，通过map[String]结构｛｝实现，以最大限度地减少内存消耗.
type String map[string]Empty

// NewString 从值列表中创建字符串.
func NewString(items ...string) String {
	ss := String{}
	ss.Insert(items...)
	return ss
}

// StringKeySet 从映射[String]（？扩展 interface{} ）的键创建一个String.
// 如果传入的值实际上不是一个映射，这将导致 panic.
func StringKeySet(theMap interface{}) String {
	v := reflect.ValueOf(theMap)
	ret := String{}

	for _, keyValue := range v.MapKeys() {
		ret.Insert(keyValue.Interface().(string))
	}
	return ret
}

// Insert 将items添加到 set.
func (s String) Insert(items ...string) String {
	for _, item := range items {
		s[item] = Empty{}
	}
	return s
}

// Delete 从集合中删除所有项目.
func (s String) Delete(items ...string) String {
	for _, item := range items {
		delete(s, item)
	}
	return s
}

// Has 当且仅当集合中包含项时，Has返回true.
func (s String) Has(item string) bool {
	_, contained := s[item]
	return contained
}

// HasAll 当且仅当所有项都包含在集合中时返回true.
func (s String) HasAll(items ...string) bool {
	for _, item := range items {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// HasAny 如果集合中包含任何项，则返回true.
func (s String) HasAny(items ...string) bool {
	for _, item := range items {
		if s.Has(item) {
			return true
		}
	}
	return false
}

// Difference 返回一组不在 s2 中的对象
// For example:
// s = {a1, a2, a3}
// s2 = {a1, a2, a4, a5}
// s.Difference(s2) = {a3}
// s2.Difference(s) = {a4, a5}
func (s String) Difference(s2 String) String {
	result := NewString()
	for key := range s {
		if !s2.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// Union 并集返回一个新集合，该集合包含s或s2中的项.
// For example:
// s = {a1, a2}
// s2 = {a3, a4}
// s.Union(s2) = {a1, a2, a3, a4}
// s2.Union(s) = {a1, a2, a3, a4}
func (s String) Union(s2 String) String {
	result := NewString()
	for key := range s {
		result.Insert(key)
	}
	for key := range s2 {
		result.Insert(key)
	}
	return result
}

// Intersection 交集返回一个新集合，该集合包含s和s2中的项目.
// For example:
// s = {a1, a2}
// s2 = {a2, a3}
// s.Intersection(s2) = {a2}
func (s String) Intersection(s2 String) String {
	var walk, other String
	result := NewString()
	if s.Len() < s2.Len() {
		walk = s
		other = s2
	} else {
		walk = s2
		other = s
	}
	for key := range walk {
		if other.Has(key) {
			result.Insert(key)
		}
	}
	return result
}

// IsSuperset 当且仅当s是s2的超集时，IsSuperset返回true.
func (s String) IsSuperset(s2 String) bool {
	for item := range s2 {
		if !s.Has(item) {
			return false
		}
	}
	return true
}

// Equal 当且仅当s等于（作为一个集合）s2时，Equal返回true.
// 如果两个集合的成员相同，那么它们是相等的.
// 在实践中，这意味着相同的元素，顺序无关紧要
func (s String) Equal(s2 String) bool {
	return len(s) == len(s2) && s.IsSuperset(s2)
}

type sortableSliceOfString []string

func (s sortableSliceOfString) Len() int           { return len(s) }
func (s sortableSliceOfString) Less(i, j int) bool { return lessString(s[i], s[j]) }
func (s sortableSliceOfString) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// List 将内容作为已排序的字符串切片返回.
func (s String) List() []string {
	res := make(sortableSliceOfString, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.Sort(res)
	return []string(res)
}

// UnsortedList 以随机顺序返回包含内容的切片.
func (s String) UnsortedList() []string {
	res := make([]string, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	return res
}

// PopAny 返回集合中的单个元素.
func (s String) PopAny() (string, bool) {
	for key := range s {
		s.Delete(key)
		return key, true
	}
	var zeroValue string
	return zeroValue, false
}

// Len 返回集合的大小.
func (s String) Len() int {
	return len(s)
}

func lessString(lhs, rhs string) bool {
	return lhs < rhs
}
