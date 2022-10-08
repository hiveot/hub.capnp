package kvstore

import "sort"

// Map2SortedList returns the map as a list sorted by key
func Map2SortedList[V any](mapInput map[string]V) []V {
	keys := make([]string, 0, len(mapInput))
	for key := range mapInput {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	listOutput := make([]V, 0, len(keys))
	for _, key := range keys {
		listOutput = append(listOutput, mapInput[key])
	}
	return listOutput
}
