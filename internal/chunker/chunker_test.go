package chunker

import (
	"errors"
	"reflect"
	"testing"
)

func TestSplitWithOverlap(t *testing.T) {
	c, err := New(5, 2)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	got := c.Split("abcdefghij")
	want := []string{
		"abcde",
		"defgh",
		"ghij",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Split() = %#v, want %#v", got, want)
	}
}

func TestSplitSupportsChinese(t *testing.T) {
	c, err := New(4, 1)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	got := c.Split("检索增强生成系统")
	want := []string{
		"检索增强",
		"强生成系",
		"系统",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Split = %#v, want %#v", got, want)
	}
}

func TestSplitEmptyText(t *testing.T) {
	c, err := New(5, 1)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	got := c.Split("   ")
	if len(got) != 0 {
		t.Fatalf("expected no chunks, got %#v", got)
	}
}

func TestNewRejectsInvalidSize(t *testing.T) {
	_, err := New(0, 0)
	if !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("expected ReeInvalidSize, got %v", err)
	}
}

func TestNewRejectsInvalidOverlap(t *testing.T) {
	_, err := New(5, 5)
	if !errors.Is(err, ErrInvalidOverlap) {
		t.Fatalf("expected ErrInvalidOverlap, got %v", err)
	}
}
