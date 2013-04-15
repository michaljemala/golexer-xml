package lexer

import (
	"fmt"
	"testing"
)

func TestEmpty(t *testing.T) {
	fmt.Printf("\nTest 1 - Empty file...\n")
	_, c := NewLexer("")

	if i := <-c; i.typ != tokenError {
		t.Fatalf("Expected '%v' token, got '%v'", tokenError, i)
	}

	fmt.Printf("...OK\n")
}

var singleTag_expected []item = []item{
	item{tokenTagBegin, "<"},
	item{tokenTagName, "person"},
	item{tokenTagEndDash, "/>"},
}

func TestSingleTag(t *testing.T) {
	fmt.Print("\nTest 2 - Single tag...\n")
	_, c := NewLexer("<person/>")

	for _, e := range singleTag_expected {
		if a := <-c; a != e {
			t.Fatalf("Expected '%v', got '%v'", e, a)
		}
	}

	fmt.Printf("...OK\n")
}

func TestSingleTag_WithSpaces(t *testing.T) {
	fmt.Printf("\nTest 3 - Single tag with spaces...\n")
	_, c := NewLexer("<person   />")

	for _, e := range singleTag_expected {
		if a := <-c; a != e {
			t.Fatalf("Expected '%v', got '%v'", e, a)
		}
	}

	fmt.Printf("...OK\n")
}
