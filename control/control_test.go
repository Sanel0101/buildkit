package control

import (
	"testing"

	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/stretchr/testify/require"
)

func TestDuplicateCacheOptions(t *testing.T) {
	var testCases = []struct {
		name     string
		opts     []*controlapi.CacheOptionsEntry
		expected []*controlapi.CacheOptionsEntry
	}{
		{
			name: "avoids unique opts",
			opts: []*controlapi.CacheOptionsEntry{
				{
					Type: "registry",
					Attrs: map[string]string{
						"ref": "example.com/ref:v1.0.0",
					},
				},
				{
					Type: "local",
					Attrs: map[string]string{
						"dest": "/path/for/export",
					},
				},
			},
			expected: nil,
		},
		{
			name: "finds duplicate opts",
			opts: []*controlapi.CacheOptionsEntry{
				{
					Type: "registry",
					Attrs: map[string]string{
						"ref": "example.com/ref:v1.0.0",
					},
				},
				{
					Type: "registry",
					Attrs: map[string]string{
						"ref": "example.com/ref:v1.0.0",
					},
				},
				{
					Type: "local",
					Attrs: map[string]string{
						"dest": "/path/for/export",
					},
				},
				{
					Type: "local",
					Attrs: map[string]string{
						"dest": "/path/for/export",
					},
				},
			},
			expected: []*controlapi.CacheOptionsEntry{
				{
					Type: "registry",
					Attrs: map[string]string{
						"ref": "example.com/ref:v1.0.0",
					},
				},
				{
					Type: "local",
					Attrs: map[string]string{
						"dest": "/path/for/export",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := findDuplicateCacheOptions(tc.opts)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.expected, result)
		})
	}
}

func TestParseCacheExportIgnoreError(t *testing.T) {
	tests := map[string]struct {
		expectedIgnoreError bool
		expectedSupported   bool
	}{
		"": {
			expectedIgnoreError: false,
			expectedSupported:   false,
		},
		".": {
			expectedIgnoreError: false,
			expectedSupported:   false,
		},
		"fake": {
			expectedIgnoreError: false,
			expectedSupported:   false,
		},
		"true": {
			expectedIgnoreError: true,
			expectedSupported:   true,
		},
		"True": {
			expectedIgnoreError: true,
			expectedSupported:   true,
		},
		"TRUE": {
			expectedIgnoreError: true,
			expectedSupported:   true,
		},
		"truee": {
			expectedIgnoreError: false,
			expectedSupported:   false,
		},
		"false": {
			expectedIgnoreError: false,
			expectedSupported:   true,
		},
		"False": {
			expectedIgnoreError: false,
			expectedSupported:   true,
		},
		"FALSE": {
			expectedIgnoreError: false,
			expectedSupported:   true,
		},
		"ffalse": {
			expectedIgnoreError: false,
			expectedSupported:   false,
		},
	}

	for ignoreErrStr, test := range tests {
		t.Run(ignoreErrStr, func(t *testing.T) {
			ignoreErr, supported := parseCacheExportIgnoreError(ignoreErrStr)
			t.Log("checking expectedIgnoreError")
			require.Equal(t, test.expectedIgnoreError, ignoreErr)
			t.Log("checking expectedSupported")
			require.Equal(t, test.expectedSupported, supported)
		})
	}
}

func TestTranslateLegacySolveRequest(t *testing.T) {
	t.Run("prevents duplicate cache exports", func(t *testing.T) {
		req := &controlapi.SolveRequest{
			Cache: &controlapi.CacheOptions{
				ExportRefDeprecated: "example.com/cache:latest",
				Exports: []*controlapi.CacheOptionsEntry{
					{
						Type:  "registry",
						Attrs: map[string]string{"ref": "example.com/cache:latest"},
					},
				},
			},
		}

		translateLegacySolveRequest(req)

		// Should not add duplicate entry
		require.Len(t, req.Cache.Exports, 1)
		require.Equal(t, "example.com/cache:latest", req.Cache.Exports[0].Attrs["ref"])
	})

	t.Run("prevents duplicate cache imports", func(t *testing.T) {
		req := &controlapi.SolveRequest{
			Cache: &controlapi.CacheOptions{
				ImportRefsDeprecated: []string{"example.com/cache:v1", "example.com/cache:v1"},
				Imports: []*controlapi.CacheOptionsEntry{
					{
						Type:  "registry",
						Attrs: map[string]string{"ref": "example.com/cache:v1"},
					},
				},
			},
		}

		translateLegacySolveRequest(req)

		// Should only have one entry for each unique ref
		require.Len(t, req.Cache.Imports, 1)
		require.Equal(t, "example.com/cache:v1", req.Cache.Imports[0].Attrs["ref"])
	})

	t.Run("allows different cache refs", func(t *testing.T) {
		req := &controlapi.SolveRequest{
			Cache: &controlapi.CacheOptions{
				ExportRefDeprecated: "example.com/cache:v2",
				Exports: []*controlapi.CacheOptionsEntry{
					{
						Type:  "registry",
						Attrs: map[string]string{"ref": "example.com/cache:v1"},
					},
				},
			},
		}

		translateLegacySolveRequest(req)

		// Should have both entries since they're different
		require.Len(t, req.Cache.Exports, 2)
	})
}
