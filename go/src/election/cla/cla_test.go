package main

import (
	"testing"
)

func TestRegisterUser(t *testing.T) {
	input := []Registration{Registration{"a", "1"},
		Registration{"b", "1"},
		Registration{"a", "asdldkfj"},
		Registration{"b", "1ladsfjabsmmmzvz"},
	}
	want := []int{200, 403, 403, 403}
	for i, reg := range input {
		if status, res := RegisterUser(reg); status != want[i] || res == nil {
			t.Errorf("Error: RegisterUser(%+v) = (%d, %s)", reg, status, res)
		}
	}
}
