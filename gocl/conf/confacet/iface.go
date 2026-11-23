package confacet

// Set provides access to default value of configuration facet.
type Set interface {
	isFacetSet()

	Facets() *getter
	Tune() *tuner
}
