package database

import (
	"testing"
)

func TestBuildDSN(t *testing.T) {
	t.Run("with explicit args", func(t *testing.T) {
		dsn := BuildDSN("myhost", "5433", "mydb", "myuser", "mypass", "require")
		expected := "host=myhost port=5433 dbname=mydb user=myuser password=mypass sslmode=require"
		if dsn != expected {
			t.Errorf("expected %q, got %q", expected, dsn)
		}
	})

	t.Run("with empty args uses defaults", func(t *testing.T) {
		dsn := BuildDSN("", "", "", "", "", "")
		expected := "host=localhost port=5432 dbname=napkin_notes user=postgres password= sslmode=disable"
		if dsn != expected {
			t.Errorf("expected %q, got %q", expected, dsn)
		}
	})
}
