package rest

func filled(v ...string) bool {
	for _, v := range v {
		if v == "" {
			return false
		}
	}
	return true
}

func notFilled(v ...string) bool {
	return !filled(v...)
}

func emptyBody() []byte {
	return []byte{}
}