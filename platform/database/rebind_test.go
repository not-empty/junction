package database

import "testing"

func TestRebindDollar(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "no placeholders",
			query: "SELECT 1",
			want:  "SELECT 1",
		},
		{
			name:  "numbers placeholders in order",
			query: "UPDATE t SET name = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
			want:  "UPDATE t SET name = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL",
		},
		{
			name:  "keeps ? inside a string literal",
			query: "SELECT id FROM t WHERE name = '?' AND id = ?",
			want:  "SELECT id FROM t WHERE name = '?' AND id = $1",
		},
		{
			name:  "keeps ? inside a literal with an escaped quote",
			query: "SELECT id FROM t WHERE name = 'it''s ?' AND id = ?",
			want:  "SELECT id FROM t WHERE name = 'it''s ?' AND id = $1",
		},
		{
			name:  "keeps ? inside a quoted identifier",
			query: `SELECT "a?" FROM t WHERE id = ?`,
			want:  `SELECT "a?" FROM t WHERE id = $1`,
		},
		{
			name:  "keeps ? inside comments",
			query: "SELECT id -- why?\nFROM t /* or? */ WHERE id = ?",
			want:  "SELECT id -- why?\nFROM t /* or? */ WHERE id = $1",
		},
		{
			name:  "?? writes a literal ?",
			query: "SELECT id FROM t WHERE data ?? 'key' AND id = ?",
			want:  "SELECT id FROM t WHERE data ? 'key' AND id = $1",
		},
		{
			name:  "unterminated literal is kept as is",
			query: "SELECT id FROM t WHERE id = ? AND name = '?",
			want:  "SELECT id FROM t WHERE id = $1 AND name = '?",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rebindDollar(tc.query)
			if got != tc.want {
				t.Errorf("rebindDollar(%q)\n got %q\nwant %q", tc.query, got, tc.want)
			}
		})
	}
}
