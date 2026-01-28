package aggregator

func copySpecKeys(specKeys map[string]string) map[string]string {
	if len(specKeys) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(specKeys))
	for key, value := range specKeys {
		out[key] = value
	}
	return out
}
