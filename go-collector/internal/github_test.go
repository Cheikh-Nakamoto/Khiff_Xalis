package internal

import (
	"testing"
	"time"
)

func TestParseCSV_ValidData(t *testing.T) {
	csv := `date,open,high,low,close,volume
2024-01-02,1200.00,1250.00,1180.00,1230.00,50000
2024-01-03,1230.00,1270.00,1210.00,1260.00,75000
2024-01-04,1260.00,1300.00,1240.00,1285.00,62000
`

	records, err := parseCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	// Verify first record.
	r := records[0]
	expectedDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if !r.Date.Equal(expectedDate) {
		t.Errorf("date: got %v, want %v", r.Date, expectedDate)
	}
	if r.Open != 1200.00 {
		t.Errorf("open: got %f, want 1200.00", r.Open)
	}
	if r.High != 1250.00 {
		t.Errorf("high: got %f, want 1250.00", r.High)
	}
	if r.Low != 1180.00 {
		t.Errorf("low: got %f, want 1180.00", r.Low)
	}
	if r.Close != 1230.00 {
		t.Errorf("close: got %f, want 1230.00", r.Close)
	}
	if r.Volume != 50000 {
		t.Errorf("volume: got %d, want 50000", r.Volume)
	}

	// Verify last record.
	r2 := records[2]
	if r2.Close != 1285.00 {
		t.Errorf("close[2]: got %f, want 1285.00", r2.Close)
	}
	if r2.Volume != 62000 {
		t.Errorf("volume[2]: got %d, want 62000", r2.Volume)
	}
}

func TestParseCSV_EmptyInput(t *testing.T) {
	records, err := parseCSV("")
	if err != nil {
		t.Fatalf("unexpected error on empty input: %v", err)
	}
	if records != nil {
		t.Errorf("expected nil for empty input, got %d records", len(records))
	}
}

func TestParseCSV_HeaderOnly(t *testing.T) {
	csv := "date,open,high,low,close,volume\n"

	records, err := parseCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records for header-only, got %d", len(records))
	}
}

func TestParseCSV_MalformedRows(t *testing.T) {
	csv := `date,open,high,low,close,volume
2024-01-02,1200.00,1250.00,1180.00,1230.00,50000
bad-date,1200,1250,1180,1230,50000
2024-01-03,not_a_number,1270,1210,1260,75000
2024-01-04,1260,1300
2024-01-05,1260.00,1300.00,1240.00,1285.00,62000
`

	records, err := parseCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only row 1 and row 5 should parse successfully.
	// Row 2: bad date, Row 3: "not_a_number" in open, Row 4: only 6 fields but < 6 columns.
	if len(records) != 2 {
		t.Fatalf("expected 2 valid records, got %d", len(records))
	}

	if records[0].Close != 1230.00 {
		t.Errorf("record[0] close: got %f, want 1230.00", records[0].Close)
	}
	if records[1].Close != 1285.00 {
		t.Errorf("record[1] close: got %f, want 1285.00", records[1].Close)
	}
}

func TestParseCSV_AlternativeDateFormats(t *testing.T) {
	csv := `date,open,high,low,close,volume
02/01/2024,1200,1250,1180,1230,50000
2024/03/15,1300,1350,1280,1320,80000
`

	records, err := parseCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// 02/01/2024 -> January 2, 2024 (dd/mm/yyyy format).
	expectedDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	if !records[0].Date.Equal(expectedDate) {
		t.Errorf("date[0]: got %v, want %v", records[0].Date, expectedDate)
	}

	// 2024/03/15 -> March 15, 2024 (yyyy/mm/dd format).
	expectedDate2 := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	if !records[1].Date.Equal(expectedDate2) {
		t.Errorf("date[1]: got %v, want %v", records[1].Date, expectedDate2)
	}
}

func TestParseCSV_DecimalVolume(t *testing.T) {
	// Some sources provide volume as float (e.g. "50000.0").
	csv := `date,open,high,low,close,volume
2024-01-02,1200,1250,1180,1230,50000.0
`

	records, err := parseCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "50000.0" is not a valid int64, so this row should be skipped.
	if len(records) != 0 {
		t.Errorf("expected 0 records for decimal volume, got %d", len(records))
	}
}

func TestParseCSV_Whitespace(t *testing.T) {
	csv := `date,open,high,low,close,volume
 2024-01-02 , 1200 , 1250 , 1180 , 1230 , 50000
`

	records, err := parseCSV(csv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record with whitespace trimming, got %d", len(records))
	}

	if records[0].Open != 1200.00 {
		t.Errorf("open with whitespace: got %f, want 1200.00", records[0].Open)
	}
	if records[0].Volume != 50000 {
		t.Errorf("volume with whitespace: got %d, want 50000", records[0].Volume)
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input string
		want  time.Time
		err   bool
	}{
		{"2024-01-02", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{"02/01/2024", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{"2024/03/15", time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), false},
		{"not-a-date", time.Time{}, true},
		{"", time.Time{}, true},
	}

	for _, tt := range tests {
		got, err := parseDate(tt.input)
		if tt.err && err == nil {
			t.Errorf("parseDate(%q): expected error, got %v", tt.input, got)
		}
		if !tt.err && err != nil {
			t.Errorf("parseDate(%q): unexpected error: %v", tt.input, err)
		}
		if !tt.err && !got.Equal(tt.want) {
			t.Errorf("parseDate(%q): got %v, want %v", tt.input, got, tt.want)
		}
	}
}
