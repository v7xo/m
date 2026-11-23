package confacet

type set struct {
	vals map[Key]any
}

var _ Set = &set{}

func (s *set) isFacetSet() {}

func (s *set) Facets() *getter { return &getter{set: s} }
func (s *set) Tune() *tuner    { return newFrom(s) }
