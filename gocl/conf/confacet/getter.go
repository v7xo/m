package confacet

type getter struct{ set *set }

func (g *getter) Val[V any](k Key, def V) V {
	existing, ok := g.set.vals[k]
	if !ok {
		return def
	}

	if typed, ok := existing.(V); ok {
		return typed
	}

	return def
}
