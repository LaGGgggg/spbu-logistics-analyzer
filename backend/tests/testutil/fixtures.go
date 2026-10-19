// Package testutil contains helpers shared only by backend tests.
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// FixtureDir locates fixtures independently of the test package nesting depth.
func FixtureDir(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			fixtures := filepath.Join(dir, "tests", "testdata")
			if _, err := os.Stat(fixtures); err != nil {
				t.Fatal(err)
			}
			return fixtures
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("backend module root not found")
		}
		dir = parent
	}
}

func FixturePath(t testing.TB, name string) string {
	t.Helper()
	path := filepath.Join(FixtureDir(t), name)
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	return path
}
