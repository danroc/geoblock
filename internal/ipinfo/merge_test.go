package ipinfo

import "testing"

func TestMergeResolutions(t *testing.T) {
	tests := []struct {
		name        string
		resolutions []Resolution
		want        Resolution
	}{
		{
			name:        "empty slice",
			resolutions: []Resolution{},
			want:        Resolution{},
		},
		{
			name: "single resolution",
			resolutions: []Resolution{
				{CountryCode: "US", ASN: 12345, Organization: "TestOrg"},
			},
			want: Resolution{CountryCode: "US", ASN: 12345, Organization: "TestOrg"},
		},
		{
			name: "merge country and ASN",
			resolutions: []Resolution{
				{CountryCode: "US"},
				{ASN: 12345, Organization: "TestOrg"},
			},
			want: Resolution{CountryCode: "US", ASN: 12345, Organization: "TestOrg"},
		},
		{
			name: "last non-zero wins",
			resolutions: []Resolution{
				{CountryCode: "US", ASN: 111},
				{CountryCode: "FR", ASN: 222},
			},
			want: Resolution{CountryCode: "FR", ASN: 222},
		},
		{
			name: "zero values do not overwrite",
			resolutions: []Resolution{
				{CountryCode: "US", ASN: 12345, Organization: "First"},
				{Organization: "Second"},
				{ASN: 0}, // Zero should not overwrite
			},
			want: Resolution{CountryCode: "US", ASN: 12345, Organization: "Second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeResolutions(tt.resolutions)
			if got != tt.want {
				t.Errorf("mergeResolutions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
