package main

import "testing"

func TestMacrosFromPayload(t *testing.T) {
	// Ctrl+C as ratbagd reports it: press ctrl, press c, release c, release ctrl.
	payload := [][]interface{}{
		{uint32(1), uint32(29)},
		{uint32(1), uint32(46)},
		{uint32(2), uint32(46)},
		{uint32(2), uint32(29)},
	}
	want := []MacroEvent{
		{Type: 1, Keycode: 29},
		{Type: 1, Keycode: 46},
		{Type: 2, Keycode: 46},
		{Type: 2, Keycode: 29},
	}

	got := macrosFromPayload(payload)
	if len(got) != len(want) {
		t.Fatalf("got %d events, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("event %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestMacrosFromPayloadSkipsMalformed(t *testing.T) {
	payload := [][]interface{}{
		{uint32(1)}, // not a pair
		{uint32(1), uint32(30)},
	}

	got := macrosFromPayload(payload)
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	if got[0].Type != 1 || got[0].Keycode != 30 {
		t.Errorf("got %+v, want {Type:1 Keycode:30}", got[0])
	}
}
