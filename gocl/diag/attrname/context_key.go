package attrname

// ContextKey represents unified type of value denoting context scoped value key.
type ContextKey string

const ContextKeyLogAttrs = ContextKey(LogAttrs)

func (v ContextKey) String() string { return string(v) }
