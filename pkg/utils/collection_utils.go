package collectionutils

// MergeMaps Merges two maps and return a new one with the keys of both.
// In case of key collision the second map key will be kept.
func MergeMaps[K comparable, V any](map1 map[K]V, map2 map[K]V) map[K]V {
	mergeMap := make(map[K]V)
	if map1 != nil {
		for key, val := range map1 {
			mergeMap[key] = val
		}
	}
	if map2 != nil {
		for key, val := range map2 {
			mergeMap[key] = val
		}
	}

	return mergeMap
}
