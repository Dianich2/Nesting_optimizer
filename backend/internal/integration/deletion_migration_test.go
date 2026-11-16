package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhysicalDeletionMigration(t *testing.T) {
	db := newTestSchema(t)
	files, err := filepath.Glob("../../migrations/*.sql")
	if err != nil || len(files) == 0 {
		t.Fatalf("migrations: %v", err)
	}
	exec := func(sql string) {
		t.Helper()
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}
	count := func(sql string, want int) {
		t.Helper()
		var got int
		if err := db.Get(&got, sql); err != nil || got != want {
			t.Fatalf("%s: got %d, want %d (%v)", sql, got, want, err)
		}
	}
	var up, down string
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		parts := strings.SplitN(string(raw), "-- +goose Down", 2)
		if strings.HasSuffix(path, "use_physical_deletion.sql") {
			up, down = parts[0], parts[1]
		} else {
			exec(parts[0])
		}
	}
	if up == "" {
		t.Fatal("migration not found")
	}
	exec(`
  INSERT INTO users(id,login,email,password_hash,first_name,last_name,deleted_at)
  VALUES (1,'active','a@test','hash','Active','User',NULL),(2,'removed','r@test','hash','Removed','User',now());
  INSERT INTO projects(id,user_id,name,deleted_at)
  VALUES (1,1,'Active',NULL),(2,2,'Hidden by account',NULL),(3,1,'Removed',now());
  INSERT INTO surfaces(id,user_id,name,geometry,deleted_at)
  SELECT id,owner,label,ST_GeomFromText('POLYGON((0 0,10 0,10 10,0 10,0 0))'),deleted
  FROM (VALUES (1,1,'Active',NULL::timestamptz),(2,1,'Removed',now()),(3,2,'Hidden by account',NULL)) AS v(id,owner,label,deleted);
  INSERT INTO patterns(id,user_id,name,geometry) SELECT 1,1,'Pattern',geometry FROM surfaces WHERE id=1;
  INSERT INTO project_surfaces(id,project_id,source_surface_id,name,geometry)
  SELECT id,project_id,2,'Copy',geometry FROM (VALUES(1,1),(2,2),(3,3)) AS v(id,project_id) CROSS JOIN (SELECT geometry FROM surfaces WHERE id=2) AS s;
  INSERT INTO project_patterns(id,project_id,source_pattern_id,name,geometry) SELECT 1,1,1,'Copy',geometry FROM patterns;
  INSERT INTO placements(id,project_surface_id,project_pattern_id,x,y,rotation,deleted_at) VALUES(1,1,1,5,5,0,NULL),(2,1,1,5,5,0,now());
 `)
	exec(up)
	columns := "SELECT count(*) FROM information_schema.columns WHERE table_schema=current_schema() AND column_name='deleted_at'"
	count(columns, 0)
	count("SELECT count(*) FROM pg_indexes WHERE schemaname=current_schema() AND indexdef LIKE '%deleted_at%'", 0)
	count("SELECT count(*) FROM projects", 1)
	count("SELECT count(*) FROM surfaces", 1)
	count("SELECT count(*) FROM placements", 1)
	count("SELECT count(*) FROM project_surfaces WHERE id=1 AND source_surface_id IS NULL AND ST_Area(geometry)=100", 1)
	exec(down)
	count(columns, 7)
	count("SELECT count(*) FROM pg_indexes WHERE schemaname=current_schema() AND indexdef LIKE '%deleted_at%'", 7)
	exec(up)
	count(columns, 0)
	count("SELECT count(*) FROM users", 1)
	count("SELECT count(*) FROM project_surfaces", 1)
	exec("DELETE FROM users WHERE id=1")
	for _, table := range []string{"users", "projects", "surfaces", "patterns", "project_surfaces", "project_patterns", "placements"} {
		count("SELECT count(*) FROM "+table, 0)
	}
}
