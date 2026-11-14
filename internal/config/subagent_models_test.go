package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubagentModelAllowlist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []SubagentModel
	}{
		{
			name: "unset",
		},
		{
			name: "provider and model",
			in:   []string{"openai/gpt-5-mini"},
			want: []SubagentModel{{Provider: "openai", Model: "gpt-5-mini"}},
		},
		{
			name: "model keeps its own slashes",
			in:   []string{"openrouter/anthropic/claude-haiku-4.5"},
			want: []SubagentModel{{Provider: "openrouter", Model: "anthropic/claude-haiku-4.5"}},
		},
		{
			name: "surrounding space is trimmed",
			in:   []string{"  mock / mock-model  "},
			want: []SubagentModel{{Provider: "mock", Model: "mock-model"}},
		},
		{
			name: "malformed entries are dropped",
			in:   []string{"gpt-5-mini", "openai/", "/mock-model", "openai/gpt-5-mini"},
			want: []SubagentModel{{Provider: "openai", Model: "gpt-5-mini"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			opts := &Options{SubagentModels: tt.in}
			require.Equal(t, tt.want, opts.SubagentModelAllowlist())
		})
	}
}

func TestSubagentModelAllowlistNilOptions(t *testing.T) {
	t.Parallel()
	var opts *Options
	require.Empty(t, opts.SubagentModelAllowlist())
	require.NoError(t, opts.ValidateSubagentModels())
}

func TestValidateSubagentModels(t *testing.T) {
	t.Parallel()

	require.NoError(t, (&Options{
		SubagentModels: []string{"openai/gpt-5-mini", "openrouter/anthropic/claude-haiku-4.5"},
	}).ValidateSubagentModels())

	for _, entry := range []string{"gpt-5-mini", "openai/", "/mock-model", ""} {
		err := (&Options{SubagentModels: []string{entry}}).ValidateSubagentModels()
		require.ErrorContains(t, err, "is not provider/model", "entry %q", entry)
	}
}
