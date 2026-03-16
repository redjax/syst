package terminal

import "testing"

func TestResponsiveTUIHelper_Defaults(t *testing.T) {
	h := NewResponsiveTUIHelper()

	w, ht := h.GetSize()
	if w != 80 || ht != 24 {
		t.Errorf("default size = (%d, %d), want (80, 24)", w, ht)
	}
}

func TestResponsiveTUIHelper_SetSize(t *testing.T) {
	h := NewResponsiveTUIHelper()
	h.SetSize(120, 40)

	if h.GetWidth() != 120 {
		t.Errorf("width = %d, want 120", h.GetWidth())
	}
	if h.GetHeight() != 40 {
		t.Errorf("height = %d, want 40", h.GetHeight())
	}
}

func TestResponsiveTUIHelper_GetContentWidth(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		wantMin40 bool
	}{
		{"wide terminal", 120, true},
		{"narrow terminal", 30, true},
		{"minimal terminal", 10, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewResponsiveTUIHelper()
			h.SetSize(tt.width, 24)
			got := h.GetContentWidth()
			if got < 40 {
				t.Errorf("GetContentWidth() = %d, want >= 40", got)
			}
		})
	}
}

func TestResponsiveTUIHelper_CalculateBarLength(t *testing.T) {
	h := NewResponsiveTUIHelper()
	h.SetSize(100, 24)

	got := h.CalculateBarLength(10, 50)
	if got < 10 || got > 50 {
		t.Errorf("CalculateBarLength(10, 50) = %d, want 10-50", got)
	}

	// Narrow terminal
	h.SetSize(20, 24)
	got = h.CalculateBarLength(10, 50)
	if got < 10 {
		t.Errorf("CalculateBarLength with narrow terminal = %d, want >= 10", got)
	}
}

func TestResponsiveTUIHelper_CalculateMaxItemsForHeight(t *testing.T) {
	tests := []struct {
		name          string
		height        int
		linesPerItem  int
		reservedLines int
		wantMin       int
	}{
		{"normal", 40, 2, 5, 1},
		{"tiny height", 5, 3, 10, 1},
		{"exact fit", 20, 2, 0, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewResponsiveTUIHelper()
			h.SetSize(80, tt.height)
			got := h.CalculateMaxItemsForHeight(tt.linesPerItem, tt.reservedLines)
			if got < tt.wantMin {
				t.Errorf("CalculateMaxItemsForHeight(%d, %d) = %d, want >= %d",
					tt.linesPerItem, tt.reservedLines, got, tt.wantMin)
			}
		})
	}
}
