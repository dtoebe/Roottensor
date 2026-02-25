package store

import (
	"database/sql"
	"os"
	"reflect"
	"testing"
)

func TestNewSQLiteDB(t *testing.T) {
	f := testDBFileSetup(t)
	defer testDBFileCleanup(t, f)

	db, err := NewSQLiteDB(f)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	t.Run("SQLiteDB: exists", func(t *testing.T) {
		if db == nil {
			t.Fatalf("SQLiteDB is nil")
		}
		if db.db == nil {
			t.Error("SQLiteDB.db is nil")
		}
		if db.path != f {
			t.Errorf("SQLiteDB.path: got: %s; want: %s", db.path, f)
		}
	})
}

func TestSQLiteDB_Close(t *testing.T) {
	t.Run("Close: success", func(t *testing.T) {
		f := testDBFileSetup(t)
		defer testDBFileCleanup(t, f)

		db, err := NewSQLiteDB(f)
		if err != nil {
			t.Fatalf("failed to create test db: %v", err)
		}

		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}
	})

	t.Run("Close: nil db", func(t *testing.T) {
		f := testDBFileSetup(t)
		defer testDBFileCleanup(t, f)

		db, err := NewSQLiteDB(f)
		if err != nil {
			t.Fatalf("failed to create test db: %v", err)
		}
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close db: %v", err)
		}

		if err := db.Close(); err == nil {
			t.Fatal("failed to catch nil db err")
		}
	})

}

func TestSQLiteDB_Exec(t *testing.T) {
	f := testDBFileSetup(t)
	defer testDBFileCleanup(t, f)

	db, err := NewSQLiteDB(f)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	t.Run("Exec: success", func(t *testing.T) {
		res, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS test (
			id INTEGER PRIMARY KEY,
			name TEXT
		)`)

		if err != nil {
			t.Errorf("Exec gave a non nil error: %v", err)
		}

		if reflect.TypeOf(res).String() == "sql.Result" {
			t.Errorf("Exec did not give sql.Result; got: %v", reflect.TypeOf(res))
		}
	})

	t.Run("Exec: fail", func(t *testing.T) {
		if _, err := db.Exec("foo"); err == nil {
			t.Fatal("expected nil error")
		}
	})
}

func TestSQLiteDB_QueryRow(t *testing.T) {
	f := testDBFileSetup(t)
	defer testDBFileCleanup(t, f)

	db, err := NewSQLiteDB(f)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	testDBInsertDummyData(t, db)

	t.Run("QueryRow: success", func(t *testing.T) {
		var name string
		want := "Daniel"
		row := db.QueryRow("SELECT name FROM test WHERE name = ? LIMIT 1", want)
		err := row.Scan(&name)
		if err != nil {
			t.Fatalf("error receiving row: %v", err)
		}
		if name != want {
			t.Fatalf("name: got: %s; want: %s", name, want)
		}
	})

	t.Run("QueryRow: no_rows_found", func(t *testing.T) {
		var name string
		row := db.QueryRow("SELECT name FROM test WHERE name = ? LIMIT 1", "Foo")
		err := row.Scan(&name)
		if err != sql.ErrNoRows {
			t.Fatal("rows returned")
		}
	})
}

func TestSQLiteDB_Query(t *testing.T) {
	f := testDBFileSetup(t)
	defer testDBFileCleanup(t, f)

	db, err := NewSQLiteDB(f)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	defer db.Close()

	testDBInsertDummyData(t, db)

	t.Run("Query: success", func(t *testing.T) {
		rows, err := db.Query("SELECT name FROM test WHERE name LIKE ? ORDER BY name", "Jo%")
		if err != nil {
			t.Fatalf("error querying rows: %v", err)
		}
		defer rows.Close()

		var got []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			got = append(got, name)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}

		if len(got) <= 0 {
			t.Fatalf("expected at least 1 row; got: %d", len(got))
		}
	})

	t.Run("Query: no_rows_found", func(t *testing.T) {
		rows, err := db.Query("SELECT name FROM test WHERE name LIKE ? LIMIT 1", "Foo%")
		if err != nil {
			t.Fatalf("error querying rows: %v", err)
		}
		var got []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatal(err)
			}
			got = append(got, name)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}

		if len(got) >= 1 {
			t.Fatalf("expected 0 rows; got: %d", len(got))
		}
	})
}

func testDBInsertDummyData(t *testing.T, db *SQliteDB) {
	t.Helper()

	query := `
		CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY, name TEXT);
		INSERT INTO test (name) VALUES ('Daniel');
		INSERT INTO test (name) VALUES ('Jo');
		INSERT INTO test (name) VALUES ('John');
	`
	_, err := db.Exec(query)
	if err != nil {
		t.Fatalf("Helper Exec gave a non nil error: %v", err)
	}
}

func testDBFileSetup(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "rt-sqlite-*.db")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}

	f.Close()
	return f.Name()
}

func testDBFileCleanup(t *testing.T, fileName string) {
	t.Helper()
	if err := os.Remove(fileName); err != nil {
		t.Fatalf("unable to delete test db file: %v", err)
	}
}
