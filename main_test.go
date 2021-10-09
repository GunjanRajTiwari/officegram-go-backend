package main

import "testing"

func TestEncrypt(t *testing.T) {
	if encrypt("password") != encrypt("password") {
		t.Error("Expected same hash for same passwords")
	}
}