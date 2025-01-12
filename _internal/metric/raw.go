package metric

type RawMetric struct {
	Name  string
	Kind  string
	Value string
}

func NewRawMetric(name, kind, value string) RawMetric {
	return RawMetric{Name: name, Kind: kind, Value: value}
}
