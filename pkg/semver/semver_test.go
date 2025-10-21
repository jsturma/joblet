package semver

import "testing"

func TestNewVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor int
		wantMinor int
		wantPatch int
		wantErr   bool
	}{
		{
			name:      "valid version",
			version:   "1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "valid version with v prefix",
			version:   "v1.2.3",
			wantMajor: 1,
			wantMinor: 2,
			wantPatch: 3,
			wantErr:   false,
		},
		{
			name:      "version with double digits",
			version:   "1.10.0",
			wantMajor: 1,
			wantMinor: 10,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:      "version 0.0.0",
			version:   "0.0.0",
			wantMajor: 0,
			wantMinor: 0,
			wantPatch: 0,
			wantErr:   false,
		},
		{
			name:    "invalid format - too few parts",
			version: "1.2",
			wantErr: true,
		},
		{
			name:    "invalid format - too many parts",
			version: "1.2.3.4",
			wantErr: true,
		},
		{
			name:    "invalid major version",
			version: "a.2.3",
			wantErr: true,
		},
		{
			name:    "invalid minor version",
			version: "1.b.3",
			wantErr: true,
		},
		{
			name:    "invalid patch version",
			version: "1.2.c",
			wantErr: true,
		},
		{
			name:    "negative version component",
			version: "1.-2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewVersion(%q) error = %v, wantErr %v", tt.version, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.major != tt.wantMajor || got.minor != tt.wantMinor || got.patch != tt.wantPatch {
					t.Errorf("NewVersion(%q) = {%d, %d, %d}, want {%d, %d, %d}",
						tt.version, got.major, got.minor, got.patch,
						tt.wantMajor, tt.wantMinor, tt.wantPatch)
				}
			}
		})
	}
}

func TestVersion_GreaterThan(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want bool
	}{
		{
			name: "greater major version",
			v1:   "2.0.0",
			v2:   "1.9.9",
			want: true,
		},
		{
			name: "greater minor version",
			v1:   "1.10.0",
			v2:   "1.9.0",
			want: true,
		},
		{
			name: "greater patch version",
			v1:   "1.2.10",
			v2:   "1.2.9",
			want: true,
		},
		{
			name: "equal versions",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: false,
		},
		{
			name: "lesser major version",
			v1:   "1.9.9",
			v2:   "2.0.0",
			want: false,
		},
		{
			name: "lesser minor version",
			v1:   "1.9.0",
			v2:   "1.10.0",
			want: false,
		},
		{
			name: "lesser patch version",
			v1:   "1.2.9",
			v2:   "1.2.10",
			want: false,
		},
		{
			name: "1.3.10 > 1.3.9 (not string comparison)",
			v1:   "1.3.10",
			v2:   "1.3.9",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, err := NewVersion(tt.v1)
			if err != nil {
				t.Fatalf("NewVersion(%q) error = %v", tt.v1, err)
			}
			v2, err := NewVersion(tt.v2)
			if err != nil {
				t.Fatalf("NewVersion(%q) error = %v", tt.v2, err)
			}

			got := v1.GreaterThan(v2)
			if got != tt.want {
				t.Errorf("%s.GreaterThan(%s) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "simple version",
			version: "1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "version with v prefix",
			version: "v1.2.3",
			want:    "1.2.3",
		},
		{
			name:    "double digit version",
			version: "1.10.0",
			want:    "1.10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := NewVersion(tt.version)
			if err != nil {
				t.Fatalf("NewVersion(%q) error = %v", tt.version, err)
			}
			if got := v.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersion_Equal(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want bool
	}{
		{
			name: "equal versions",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: true,
		},
		{
			name: "different patch",
			v1:   "1.2.3",
			v2:   "1.2.4",
			want: false,
		},
		{
			name: "different minor",
			v1:   "1.2.3",
			v2:   "1.3.3",
			want: false,
		},
		{
			name: "different major",
			v1:   "1.2.3",
			v2:   "2.2.3",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1, _ := NewVersion(tt.v1)
			v2, _ := NewVersion(tt.v2)

			if got := v1.Equal(v2); got != tt.want {
				t.Errorf("%s.Equal(%s) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}
