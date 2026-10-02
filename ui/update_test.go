package ui

import "testing"

func TestUpdateWanted(t *testing.T) {
	if updateWanted("2.9.1", "2.9.1", false) {
		t.Fatal("same version should wait")
	}
	if !updateWanted("2.9.2", "2.9.1", false) {
		t.Fatal("newer version should be offered")
	}
	if !updateWanted("2.9.1", "2.9.1", true) {
		t.Fatal("test mode should fetch the current release")
	}
	if !updateWanted("2.0.0", "2.9.1", true) {
		t.Fatal("test mode should fetch an older release too")
	}
	if updateWanted("", "2.9.1", true) {
		t.Fatal("empty latest is not a release")
	}
}

func TestDownloadBar(t *testing.T) {
	if _, _, known := downloadBar(10, -1); known {
		t.Fatal("missing length")
	}
	if _, _, known := downloadBar(0, 0); known {
		t.Fatal("zero length")
	}
	value, maximum, known := downloadBar(0, 2000)
	if !known || value != 0 || maximum != 1000 {
		t.Fatalf("start = %d/%d known %v", value, maximum, known)
	}
	value, maximum, known = downloadBar(2000, 2000)
	if !known || value != 1000 || maximum != 1000 {
		t.Fatalf("done = %d/%d", value, maximum)
	}
	value, _, _ = downloadBar(500, 2000)
	if value != 250 {
		t.Fatalf("quarter = %d", value)
	}
}
