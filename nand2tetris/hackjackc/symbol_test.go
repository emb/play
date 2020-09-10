package main

import "testing"

func TestSymbolTable(t *testing.T) {
	// Setup
	syms := NewSymbols()
	syms.Global("a", "int", "static", 0)
	syms.Global("b", "int", "field", 1)
	syms.InitLocal()
	syms.Local("b", "string", "var", 0)
	syms.Local("d", "string", "var", 1)
	// Expectation
	ag := Symbol{Type:"int", Kind:"static"}
	bg := Symbol{Type:"int", Kind:"field", Index: 1}
	bl := Symbol{Type:"string", Kind:"var"}
	dl := Symbol{Type:"string", Kind:"var", Index: 1}

	// Check global
	agGot, ok := syms.Lookup("a")
	if !ok {
		t.Error("Lookup(a): want ok got !ok")
	}
	if agGot != ag {
		t.Errorf("Lookup(a): want %v got %v", ag, agGot)
	}
	_, ok = syms.Lookup("something-that-should-not-exist")
	if ok {
		t.Error("Lookup(something-that-should-not-exist) returns ok!")
	}

	// Check local
	blGot, ok := syms.Lookup("b")
	if !ok {
		t.Error("Lookup(b): want ok got !ok")
	}
	if blGot != bl {
		t.Errorf("Lookup(b): want %v got %v", bl, blGot)
	}
	dlGot, ok := syms.Lookup("d")
	if !ok {
		t.Error("Lookup(d): want ok got !ok")
	}
	if dlGot != dl {
		t.Errorf("Lookup(d): want %v got %v", dl, dlGot)
	}

	// Check local after init
	syms.InitLocal()
	_, ok = syms.Lookup("d")
	if ok {
		t.Errorf("Lookup(d) want !ok got ok after InitLocal()")
	}
	bgGot, ok := syms.Lookup("b")
	if !ok {
		t.Errorf("Lookup(b) want ok got !ok")
	}
	if bgGot != bg {
		t.Errorf("Lookup(b) want %v got %v", bg, bgGot)
	}
}
