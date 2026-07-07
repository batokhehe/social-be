package volunteer

import (
	"encoding/json"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

// The leadership flags must serialize as explicit true / false / null (never
// omitted), so the frontend can distinguish "assigned" from "not assigned".
func TestLeadershipFlags_JSONSerialization(t *testing.T) {
	v := Volunteer{
		ID:             12,
		IndonesianName: "Budi",
		IsHuAiLeader:   boolPtr(true),
		IsHuAiDeputy:   nil,
		IsXieLiLeader:  boolPtr(false),
		IsXieLiDeputy:  nil,
	}

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	got := string(data)

	for _, want := range []string{
		`"is_hu_ai_leader":true`,
		`"is_hu_ai_deputy":null`,
		`"is_xie_li_leader":false`,
		`"is_xie_li_deputy":null`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %s in JSON, got %s", want, got)
		}
	}
}

// toEntity (used by both Detail and List paths) must pass the nullable flags
// through untouched.
func TestToEntity_LeadershipFlags(t *testing.T) {
	row := volunteerModel{
		ID:            7,
		IsHuAiLeader:  boolPtr(true),
		IsHuAiDeputy:  nil,
		IsXieLiLeader: boolPtr(false),
		IsXieLiDeputy: nil,
	}

	got := toEntity(row)

	if got.IsHuAiLeader == nil || *got.IsHuAiLeader != true {
		t.Fatalf("is_hu_ai_leader = %v, want true", got.IsHuAiLeader)
	}
	if got.IsHuAiDeputy != nil {
		t.Fatalf("is_hu_ai_deputy = %v, want nil", got.IsHuAiDeputy)
	}
	if got.IsXieLiLeader == nil || *got.IsXieLiLeader != false {
		t.Fatalf("is_xie_li_leader = %v, want false", got.IsXieLiLeader)
	}
	if got.IsXieLiDeputy != nil {
		t.Fatalf("is_xie_li_deputy = %v, want nil", got.IsXieLiDeputy)
	}
}
