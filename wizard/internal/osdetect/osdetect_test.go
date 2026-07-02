package osdetect

import "testing"

func TestParseOS(t *testing.T) {
	tests := []struct {
		goos    string
		want    OS
		wantErr bool
	}{
		{"darwin", Darwin, false},
		{"linux", Linux, false},
		{"windows", Windows, false},
		{"plan9", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := parseOS(tt.goos)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseOS(%q): expected error, got nil", tt.goos)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseOS(%q): unexpected error: %v", tt.goos, err)
		}
		if got != tt.want {
			t.Errorf("parseOS(%q) = %q, want %q", tt.goos, got, tt.want)
		}
	}
}

func TestParseArch(t *testing.T) {
	tests := []struct {
		goarch  string
		want    Arch
		wantErr bool
	}{
		{"amd64", AMD64, false},
		{"arm64", ARM64, false},
		{"386", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := parseArch(tt.goarch)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseArch(%q): expected error, got nil", tt.goarch)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseArch(%q): unexpected error: %v", tt.goarch, err)
		}
		if got != tt.want {
			t.Errorf("parseArch(%q) = %q, want %q", tt.goarch, got, tt.want)
		}
	}
}

func TestParseOSReleaseID(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "ubuntu quoted",
			content: "NAME=\"Ubuntu\"\nID=ubuntu\nID_LIKE=debian\n",
			want:    "ubuntu",
		},
		{
			name:    "fedora",
			content: "ID=fedora\nVERSION_ID=39\n",
			want:    "fedora",
		},
		{
			name:    "double quoted id",
			content: "ID=\"rhel\"\n",
			want:    "rhel",
		},
		{
			name:    "missing id",
			content: "NAME=\"Whatever\"\n",
			want:    "",
		},
		{
			name:    "empty",
			content: "",
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseOSReleaseID(tt.content); got != tt.want {
				t.Errorf("parseOSReleaseID() = %q, want %q", got, tt.want)
			}
		})
	}
}
