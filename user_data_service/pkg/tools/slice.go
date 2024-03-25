package tools

import "sort"

func ContainsStringValue(slice []string, value string) bool {
	for _, val := range slice {
		if val == value {
			return true
		}
	}
	return false
}

func FindIntersection(arr1 []string, arr2 []string) []string {
	intersection := make([]string, 0)
	m := make(map[string]struct{}, len(arr1))

	for _, num := range arr1 {
		m[num] = struct{}{}
	}

	for _, num := range arr2 {
		_, ok := m[num]
		if ok {
			intersection = append(intersection, num)
		}
	}

	return intersection
}

func MergeArrays(arr1 []string, arr2 []string) []string {
	merged := append(arr1, arr2...)
	unique := make([]string, 0)
	m := make(map[string]struct{})

	for _, num := range merged {
		_, ok := m[num]
		if !ok {
			m[num] = struct{}{}
			unique = append(unique, num)
		}
	}

	return unique
}

func IsEqualStringArrays(arr1 []string, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	sort.Strings(arr1)
	sort.Strings(arr2)

	for i := range arr1 {
		if arr1[i] != arr2[i] {
			return false
		}
	}

	return true
}

func IsSubset(arr1 []string, arr2 []string) bool {
	set := make(map[string]struct{})

	for _, s := range arr1 {
		set[s] = struct{}{}
	}

	for _, s := range arr2 {
		_, ok := set[s]
		if !ok {
			return false
		}
	}

	return true
}

func StringArrayDifference(arr1 []string, arr2 []string) []string {
	// Сортируем оба массива
	sort.Strings(arr1)
	sort.Strings(arr2)

	var diff []string
	i, j := 0, 0

	// Проходим по обоим массивам и находим разность
	for i < len(arr1) && j < len(arr2) {
		switch {
		case arr1[i] < arr2[j]:
			diff = append(diff, arr1[i])
			i++
		case arr1[i] > arr2[j]:
			j++
		default: // arr1[i] == arr2[j]
			i++
			j++
		}
	}

	// Добавляем оставшиеся элементы из arr1, если есть
	for i < len(arr1) {
		diff = append(diff, arr1[i])
		i++
	}

	return diff
}

func CommonPrefix(str1, str2 string) string {
	minLen := min(len(str1), len(str2))
	for i := 0; i < minLen; i++ {
		if str1[i] != str2[i] {
			return str1[:i]
		}
	}
	return str1[:minLen]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
