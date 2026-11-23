package confacet

type tuner struct {
	del map[Key]struct{}
	add map[Key]any

	original *set
}

func (t *tuner) Del(k Key) *tuner {
	t.del[k] = struct{}{}
	delete(t.add, k)

	return t
}

func (t *tuner) Set(k Key, v any) *tuner {
	t.add[k] = v
	delete(t.del, k)

	return t
}

func (t *tuner) Tuned() Set {
	newVals := t.add

	if t.original != nil {
		for k, v := range t.original.vals {
			if _, ok := newVals[k]; ok {
				continue
			}

			newVals[k] = v
		}
	}

	for k := range t.del {
		delete(newVals, k)
	}

	return &set{vals: newVals}
}

func newFrom(original *set) *tuner {
	return &tuner{
		del: map[Key]struct{}{},
		add: map[Key]any{},

		original: original,
	}
}

func Tuner() *tuner { return newFrom(nil) }
