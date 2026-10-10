package migrations

import (
	"fmt"
	"io"
	"testing"

	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func TestEmbeddedMigrationsAreReadable(t *testing.T) {
	source, err := iofs.New(FS, ".")
	if err != nil {
		t.Fatalf("create embedded migration source: %v", err)
	}
	defer source.Close()

	for version := uint(1); version <= 8; version++ {
		t.Run(fmt.Sprintf("version_%d", version), func(t *testing.T) {
			// Test ReadUp
			upMigration, upIdent, err := source.ReadUp(version)
			if err != nil {
				t.Fatalf("read Up migration for version %d: %v", version, err)
			}
			defer upMigration.Close()

			if upIdent == "" {
				t.Fatalf("version %d Up has empty identifier", version)
			}
			if _, err := io.ReadAll(upMigration); err != nil {
				t.Fatalf("read Up migration content for version %d: %v", version, err)
			}

			// Test ReadDown
			downMigration, downIdent, err := source.ReadDown(version)
			if err != nil {
				t.Fatalf("read Down migration for version %d: %v", version, err)
			}
			defer downMigration.Close()

			if downIdent == "" {
				t.Fatalf("version %d Down has empty identifier", version)
			}
			if _, err := io.ReadAll(downMigration); err != nil {
				t.Fatalf("read Down migration content for version %d: %v", version, err)
			}
		})
	}
}

func TestEmbeddedMigrationSixIsReadable(t *testing.T) {
	source, err := iofs.New(FS, ".")
	if err != nil {
		t.Fatalf("create embedded migration source: %v", err)
	}
	defer source.Close()

	migration, identifier, err := source.ReadUp(6)
	if err != nil {
		t.Fatalf("read migration version 6: %v", err)
	}
	defer migration.Close()

	if identifier == "" {
		t.Fatal("migration version 6 has no identifier")
	}
	if _, err := io.ReadAll(migration); err != nil {
		t.Fatalf("read migration version 6 contents: %v", err)
	}
}

func TestEmbeddedMigrationEightIsReadable(t *testing.T) {
	source, err := iofs.New(FS, ".")
	if err != nil {
		t.Fatalf("create embedded migration source: %v", err)
	}
	defer source.Close()

	migration, identifier, err := source.ReadUp(8)
	if err != nil {
		t.Fatalf("read migration version 8: %v", err)
	}
	defer migration.Close()

	if identifier == "" {
		t.Fatal("migration version 8 has no identifier")
	}
	if _, err := io.ReadAll(migration); err != nil {
		t.Fatalf("read migration version 8 contents: %v", err)
	}
}
