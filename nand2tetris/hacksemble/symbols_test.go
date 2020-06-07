package main

import "testing"

func TestSymbols(t *testing.T) {
	var s Symbols

	addr, ok := s.Get("foo")
	if ok != false {
		t.Error("Get(foo) want false got true")
	}
	if addr != 0 {
		t.Errorf("Get(foo) want 0 got %#b", addr)
	}

	s.Add("bar", 666)
	addr, ok = s.Get("bar")
	if ok == false {
		t.Error("Get(bar) want true got false")
	}
	if addr != 666 {
		t.Errorf("Get(bar) want 0b1010011010 got %#b ", addr)
	}
}

func TestReadSymbols(t *testing.T) {
	insts := []*instruction{
		newInstruction("@56", 0, 0),
		newInstruction("@i", 1, 1),
		newInstruction("(EEK)", 2, 2),
		newInstruction("@j", 3, 3),
	}

	s := readSymbols(insts)

	// Check for predefined symbol
	addr, ok := s.Get("R4")
	if !ok {
		t.Error("did not find R0 in symbol table")
	}
	if addr != 4 {
		t.Errorf("symbol R4 want 4 got %d", addr)
	}

	// Ensure we don't have numbers in symbol table
	_, ok = s.Get("56")
	if ok {
		t.Error("did not expect to find label 56")
	}

	// Ensure labels are stored with their addresses
	addr, ok = s.Get("EEK")
	if !ok {
		t.Error(`did not find label "EEK"`)
	}
	if addr != 2 {
		t.Errorf("Get(EEK) want 2 got %d", addr)
	}

	// ensure i,j has the appropriate variable address
	i, _ := s.Get("i")
	if i != 16 {
		t.Errorf("Get(i) want 16 got %d", i)
	}
	j, _ := s.Get("j")
	if j != 17 {
		t.Errorf("Get(j) want 17 got %d", j)
	}
}
