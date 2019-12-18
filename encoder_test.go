package main

import "testing"

func TestHasher(t *testing.T) {
	expected := "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="
	result, _ := CalculateHash("angryMonkey")
	if expected != result {
		t.Errorf("CalculateHash() = %q, but expected %q", result, expected)
	}
}
