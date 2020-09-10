package main

// Symbol describes symbol information
type Symbol struct {
	Type string
	Kind string
	Index int
}

// Symbols manages a symbols table while generating code.
type Symbols struct {
	global map[string]Symbol
	local map[string]Symbol
}

// NewSymbols constructs a new symbols table.
func NewSymbols() *Symbols {
	return &Symbols{
		global: make(map[string]Symbol),
	}
}

// Global inserts or updates a global symbol.
func (s*Symbols) Global(ident, typ, kind string, index int) {
	s.global[ident] = Symbol{
		Type: typ,
		Kind: kind,
		Index: index,
	}
}

// Local insert or updates a local symbol. Will panic if InitLocal is
// not called first.
func (s*Symbols) Local(ident, typ, kind string, index int) {
	s.local[ident] = Symbol{
		Type: typ,
		Kind: kind,
		Index: index,		
	}
}

// Lookup returns a symbol if one exist.
func (s*Symbols) Lookup(ident string) (Symbol, bool) {
	sym, ok := s.local[ident]
	if ok {
		return sym, ok
	}
	sym, ok = s.global[ident]
	return sym, ok
}

// InitLocal initializes local symbols space.
func (s*Symbols) InitLocal() {
	s.local = make(map[string]Symbol)
}

// Count counts the number of symbols of kind k
func (s *Symbols) Count(k string) int {
	var n int
	for _, v := range s.global {
		if v.Kind == k {
			n++
		}
	}
	for _, v := range s.local {
		if v.Kind == k {
			n++
		}
	}
	return n
}
