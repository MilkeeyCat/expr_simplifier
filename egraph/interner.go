package egraph

type stringIdx uint32

type stringInterner struct {
	strings    []string
	stringsMap map[string]stringIdx
}

func newStringInterner() *stringInterner {
	return &stringInterner{
		stringsMap: make(map[string]stringIdx),
	}
}

func (si *stringInterner) intern(str string) stringIdx {
	if idx, ok := si.stringsMap[str]; ok {
		return idx
	}

	idx := stringIdx(len(si.strings))
	si.strings = append(si.strings, str)
	si.stringsMap[str] = idx

	return idx
}

func (si *stringInterner) get(idx stringIdx) string {
	return si.strings[idx]
}
