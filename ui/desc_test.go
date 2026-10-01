package ui

import "testing"

func TestApplyDescriptionText(t *testing.T) {
	const label = "SW-1234-5678-9012"
	base := GameState{Mode: "friendcode", Description: "kept"}

	custom := applyDescriptionText(base, "In a party", label)
	if custom.Mode != "custom" || custom.Description != "In a party" {
		t.Fatalf("custom = %+v", custom)
	}

	back := applyDescriptionText(custom, label, label)
	if back.Mode != "friendcode" || back.Description != "In a party" {
		t.Fatalf("friend code match = %+v", back)
	}

	blank := applyDescriptionText(custom, "   ", label)
	if blank.Mode != "empty" || blank.Description != "" {
		t.Fatalf("blank = %+v", blank)
	}
}
