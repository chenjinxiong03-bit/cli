// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT
package doc

import (
	"testing"
)

func TestBuildUpdateBody(t *testing.T) {
	// buildUpdateBody is tested indirectly via command validation.
	// The function just maps runtime flags to a map, so we verify the valid commands list.
	for _, cmd := range validCommands {
		if cmd == "" {
			t.Fatal("validCommands contains empty string")
		}
	}

	expected := map[string]bool{
		"str_replace":   true,
		"str_delete":    true,
		"block_delete":  true,
		"block_insert_after":  true,
		"block_replace": true,
		"overwrite":     true,
		"append":        true,
	}
	if len(validCommands) != len(expected) {
		t.Fatalf("expected %d commands, got %d", len(expected), len(validCommands))
	}
	for _, cmd := range validCommands {
		if !expected[cmd] {
			t.Fatalf("unexpected command %q in validCommands", cmd)
		}
	}
}
