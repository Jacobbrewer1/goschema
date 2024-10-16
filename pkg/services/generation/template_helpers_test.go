package generation

import "testing"

func TestStructify(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "snake_case",
			in:   "snake_case",
			want: "SnakeCase",
		},
		{
			name: "snake_case_with_underscores",
			in:   "snake_case_with_underscores",
			want: "SnakeCaseWithUnderscores",
		},
		{
			name: "snake_case_with_underscores_and_numbers_123",
			in:   "snake_case_with_underscores_and_numbers_123",
			want: "SnakeCaseWithUnderscoresAndNumbers123",
		},
		{
			name: "snake_case_with_underscores_and_numbers_123_and_underscores",
			in:   "snake_case_with_underscores_and_numbers_123_and_underscores",
			want: "SnakeCaseWithUnderscoresAndNumbers123AndUnderscores",
		},
		{
			name: "single",
			in:   "single",
			want: "Single",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := structify(tt.in); got != tt.want {
				t.Errorf("structify() = %v, want %v", got, tt.want)
			}
		})
	}
}
