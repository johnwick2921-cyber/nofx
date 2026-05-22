package databento

import "testing"

func TestParseResolveResponse_FrontMonthNQ(t *testing.T) {
	// Real-shape response from /v0/symbology.resolve for symbols=NQ.c.0
	body := []byte(`{
		"result": {
			"NQ.c.0": [
				{"d0": "2026-05-22", "d1": "2026-06-19", "s": "NQM6"}
			]
		},
		"symbols": ["NQ.c.0"],
		"stype_in": "continuous",
		"stype_out": "raw_symbol",
		"start_date": "2026-05-22",
		"end_date": "2026-05-22",
		"partial": [],
		"not_found": []
	}`)

	got, err := parseResolveResponse(body, "NQ.c.0")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if got != "NQM6" {
		t.Errorf("got %q, want %q", got, "NQM6")
	}
}

func TestParseResolveResponse_NotFound(t *testing.T) {
	body := []byte(`{"result":{},"symbols":["NQ.c.0"],"not_found":["NQ.c.0"]}`)
	_, err := parseResolveResponse(body, "NQ.c.0")
	if err == nil {
		t.Fatal("want error when symbol not found, got nil")
	}
}
