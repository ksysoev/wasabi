package http

type ContextKey uint

const (
	NoKey ContextKey = iota
	ClientIP
	Stash
)
