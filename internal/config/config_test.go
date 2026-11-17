package config

import "testing"

func TestGet(t *testing.T) {
	t.Setenv("LB_SET", "x")
	t.Setenv("LB_EMPTY", "")
	if Get("LB_SET", "d") != "x" || Get("LB_EMPTY", "d") != "d" || Get("LB_UNSET_XYZ", "d") != "d" {
		t.Fatal("Get non rispetta valore e fallback")
	}
}
