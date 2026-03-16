package convert

import "testing"

func TestBytesToHumanReadable(t *testing.T) {
	tests := []struct {
		name  string
		bytes uint64
		want  string
	}{
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"500 bytes", 500, "500 B"},
		{"1023 bytes", 1023, "1023 B"},
		{"1 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1 MB", 1024 * 1024, "1.0 MB"},
		{"1 GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"1 TB", 1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{"2.5 GB", 2684354560, "2.5 GB"},
		{"100 MB", 104857600, "100.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToHumanReadable(tt.bytes)
			if got != tt.want {
				t.Errorf("BytesToHumanReadable(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{"plain bytes", "1024", 1024},
		{"bytes with B", "512B", 512},
		{"kilobytes K", "10K", 10 * 1024},
		{"kilobytes KB", "10KB", 10 * 1024},
		{"kilobytes KiB", "10KiB", 10 * 1024},
		{"megabytes", "5MB", 5 * 1024 * 1024},
		{"megabytes M", "5M", 5 * 1024 * 1024},
		{"megabytes MiB", "5MiB", 5 * 1024 * 1024},
		{"gigabytes", "1GB", 1024 * 1024 * 1024},
		{"gigabytes GiB", "1GiB", 1024 * 1024 * 1024},
		{"terabytes", "2TB", 2 * 1024 * 1024 * 1024 * 1024},
		{"with spaces", "  10MB  ", 10 * 1024 * 1024},
		{"lowercase", "10mb", 10 * 1024 * 1024},
		{"decimal", "1.5GB", int64(1.5 * 1024 * 1024 * 1024)},
		{"invalid unit", "10XB", 0},
		{"no number", "MB", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseByteSize(tt.input)
			if got != tt.want {
				t.Errorf("ParseByteSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
