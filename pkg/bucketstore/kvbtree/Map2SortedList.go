package kvbtree

// Map2SortedKeys returns the map orderedKeys as a list sorted by key
//func Map2SortedKeys[V any](mapInput map[string]V) []string {
//	keys := make([]string, 0, len(mapInput))
//	for key := range mapInput {
//		keys = append(keys, key)
//	}
//
//	sort.Strings(keys)
//	return keys
//}

// Map2SortedValues returns the map values as a list sorted by key
//func Map2SortedValues[V any](mapInput map[string]V) []V {
//	orderedKeys := make([]string, 0, len(mapInput))
//	for key := range mapInput {
//		orderedKeys = append(orderedKeys, key)
//	}
//	sort.Strings(orderedKeys)
//	listOutput := make([]V, 0, len(orderedKeys))
//	for _, key := range orderedKeys {
//		listOutput = append(listOutput, mapInput[key])
//	}
//	return listOutput
//}
