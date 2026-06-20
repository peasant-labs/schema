package village

import "testing"

func TestMigrationsEmbedded(t *testing.T) {
	files := []string{
		"007_session_entries.up.sql",
		"007_session_entries.down.sql",
		"008_cost_scope_visibility.up.sql",
		"008_cost_scope_visibility.down.sql",
	}
	for _, name := range files {
		content, err := Migrations.ReadFile(name)
		if err != nil {
			t.Errorf("Migrations.ReadFile(%q): %v", name, err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("Migrations.ReadFile(%q): empty content", name)
		}
	}
}
