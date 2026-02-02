package terraform

import "testing"

func TestNormalizeToTerraformName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "mygroup",
			expected: "mygroup",
		},
		{
			name:     "with dashes",
			input:    "my-group",
			expected: "my_group",
		},
		{
			name:     "with slashes",
			input:    "my-group/my-project",
			expected: "my_group_my_project",
		},
		{
			name:     "uppercase",
			input:    "My-Group",
			expected: "my_group",
		},
		{
			name:     "mixed",
			input:    "Parent-Group/Sub-Group/My-Project",
			expected: "parent_group_sub_group_my_project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeToTerraformName(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeToTerraformName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
