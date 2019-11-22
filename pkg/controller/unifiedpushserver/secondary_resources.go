package unifiedpushserver

type resources map[string][]string

func (res resources) add(k, v string) {
	if res == nil {
		// why would it be nil
		return
	}
	for _, e := range res[k] {
		if e == v {
			return
		}
	}
	res[k] = append(res[k], v)
}

func (res resources) remove(k, v string) {
	for i, e := range res[k] {
		if e == v {
			res[k][i] = res[k][len(res[k])-1]
			res[k] = res[k][:len(res[k])-1]
			break
		}
	}
}
