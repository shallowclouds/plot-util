package util

func ArrayToMap(raw []string) (m map[string]struct{}) {
	m = make(map[string]struct{})
	for _, s := range raw {
		m[s] = struct{}{}
	}

	return
}

func MapKeyToArray(m map[string]struct{}) []string {
	as := make([]string, 0, len(m))
	for k := range m {
		as = append(as, k)
	}

	return as
}
