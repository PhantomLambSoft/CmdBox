package repository

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestResolveOrderClause(t *testing.T) {
	allowed := map[string]bool{"name": true, "created_at": true}

	cases := []struct {
		name        string
		token       string
		wantClause  string
		wantErr     bool
		errContains string
	}{
		{name: "ascending", token: "name", wantClause: "name ASC"},
		{name: "descending", token: "-name", wantClause: "name DESC"},
		{name: "trims whitespace", token: "  name  ", wantClause: "name ASC"},
		{name: "empty token", token: "", wantErr: true, errContains: "empty order_by token"},
		{name: "whitespace only token", token: "   ", wantErr: true, errContains: "empty order_by token"},
		{name: "disallowed column", token: "secret", wantErr: true, errContains: "invalid order_by field: secret"},
		{name: "disallowed column descending", token: "-secret", wantErr: true, errContains: "invalid order_by field: -secret"},
		{name: "bare dash", token: "-", wantErr: true, errContains: "invalid order_by field: -"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveOrderClause(tc.token, allowed)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveOrderClause(%q) error = nil, want error", tc.token)
				}
				if tc.errContains != "" && err.Error() != tc.errContains {
					t.Fatalf("resolveOrderClause(%q) error = %q, want %q", tc.token, err.Error(), tc.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveOrderClause(%q) unexpected error = %v", tc.token, err)
			}
			if got != tc.wantClause {
				t.Fatalf("resolveOrderClause(%q) = %q, want %q", tc.token, got, tc.wantClause)
			}
		})
	}
}

func TestResolveOrdering(t *testing.T) {
	allowed := map[string]bool{"name": true, "created_at": true}

	t.Run("single token", func(t *testing.T) {
		got, err := resolveOrdering("name", allowed)
		if err != nil {
			t.Fatalf("resolveOrdering() error = %v", err)
		}
		want := []string{"name ASC"}
		if len(got) != 1 || got[0] != want[0] {
			t.Fatalf("resolveOrdering() = %v, want %v", got, want)
		}
	})

	t.Run("multiple tokens preserve order", func(t *testing.T) {
		got, err := resolveOrdering("-created_at,name", allowed)
		if err != nil {
			t.Fatalf("resolveOrdering() error = %v", err)
		}
		want := []string{"created_at DESC", "name ASC"}
		if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("resolveOrdering() = %v, want %v", got, want)
		}
	})

	t.Run("tokens with surrounding whitespace", func(t *testing.T) {
		got, err := resolveOrdering(" name , -created_at ", allowed)
		if err != nil {
			t.Fatalf("resolveOrdering() error = %v", err)
		}
		want := []string{"name ASC", "created_at DESC"}
		if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("resolveOrdering() = %v, want %v", got, want)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := resolveOrdering("", allowed)
		if err == nil || err.Error() != "empty order_by" {
			t.Fatalf("resolveOrdering(\"\") error = %v, want \"empty order_by\"", err)
		}
	})

	t.Run("only commas and whitespace", func(t *testing.T) {
		_, err := resolveOrdering(" , , ", allowed)
		if err == nil || err.Error() != "empty order_by" {
			t.Fatalf("resolveOrdering() error = %v, want \"empty order_by\"", err)
		}
	})

	t.Run("invalid token stops processing and returns error", func(t *testing.T) {
		_, err := resolveOrdering("name,bogus,created_at", allowed)
		if err == nil {
			t.Fatalf("resolveOrdering() error = nil, want error")
		}
		wantErr := "invalid order_by field: bogus"
		if err.Error() != wantErr {
			t.Fatalf("resolveOrdering() error = %q, want %q", err.Error(), wantErr)
		}
	})
}

func TestSplitCSV(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty string", input: "", want: []string{}},
		{name: "single item", input: "a", want: []string{"a"}},
		{name: "multiple items", input: "a,b,c", want: []string{"a", "b", "c"}},
		{name: "trims whitespace", input: " a , b ,c ", want: []string{"a", "b", "c"}},
		{name: "skips empty segments", input: "a,,b,,,c", want: []string{"a", "b", "c"}},
		{name: "only commas", input: ",,,", want: []string{}},
		{name: "only whitespace", input: "   ", want: []string{}},
		{name: "trailing and leading commas", input: ",a,b,", want: []string{"a", "b"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitCSV(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("splitCSV(%q) = %v, want %v", tc.input, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("splitCSV(%q) = %v, want %v", tc.input, got, tc.want)
				}
			}
		})
	}
}

func TestIsUniqueConstraintViolation(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		table  string
		column string
		want   bool
	}{
		{
			name:   "nil error",
			err:    nil,
			table:  "commands",
			column: "alias",
			want:   false,
		},
		{
			name:   "matching violation",
			err:    errors.New("UNIQUE constraint failed: commands.alias"),
			table:  "commands",
			column: "alias",
			want:   true,
		},
		{
			name:   "matching violation with extra context",
			err:    fmt.Errorf("create command: %w", errors.New("UNIQUE constraint failed: commands.alias")),
			table:  "commands",
			column: "alias",
			want:   true,
		},
		{
			name:   "wrong table",
			err:    errors.New("UNIQUE constraint failed: variables.name"),
			table:  "commands",
			column: "alias",
			want:   false,
		},
		{
			name:   "wrong column",
			err:    errors.New("UNIQUE constraint failed: commands.template"),
			table:  "commands",
			column: "alias",
			want:   false,
		},
		{
			name:   "unrelated error",
			err:    errors.New("no such table: commands"),
			table:  "commands",
			column: "alias",
			want:   false,
		},
		{
			name:   "table and column substrings present but not as table.column",
			err:    errors.New("some commands and alias mismatch"),
			table:  "commands",
			column: "alias",
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isUniqueConstraintViolation(tc.err, tc.table, tc.column)
			if got != tc.want {
				t.Fatalf("isUniqueConstraintViolation(%v, %q, %q) = %v, want %v", tc.err, tc.table, tc.column, got, tc.want)
			}
		})
	}
}

// searchItem backs the "search_items" table used to exercise searchWithRelevance.
type searchItem struct {
	ID          uint `gorm:"primaryKey"`
	Name        string
	Description string
	ProfileID   uint
}

func (searchItem) TableName() string { return "search_items" }

type searchResult struct {
	ID          uint
	Name        string
	Description string
	ProfileID   uint
	Relevance   float64
}

func setupSearchTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:query-helpers-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Fatalf("close sql db: %v", closeErr)
		}
	})

	if err := db.AutoMigrate(&searchItem{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	return db
}

func seedSearchItems(t *testing.T, db *gorm.DB, items ...searchItem) {
	t.Helper()
	for i := range items {
		if err := db.Create(&items[i]).Error; err != nil {
			t.Fatalf("seed search item %+v: %v", items[i], err)
		}
	}
}

func TestSearchWithRelevanceEmptyInputsNoop(t *testing.T) {
	db := setupSearchTestDB(t)
	seedSearchItems(t, db, searchItem{Name: "apple"})
	allowed := map[string]bool{"name": true}

	t.Run("empty query returns nil without querying", func(t *testing.T) {
		var dest []searchResult
		err := searchWithRelevance(db, "search_items", "", []string{"name"}, allowed, "id ASC", 10, "", nil, &dest)
		if err != nil {
			t.Fatalf("searchWithRelevance() error = %v", err)
		}
		if dest != nil {
			t.Fatalf("dest = %v, want nil (untouched)", dest)
		}
	})

	t.Run("no fields returns nil without querying", func(t *testing.T) {
		var dest []searchResult
		err := searchWithRelevance(db, "search_items", "apple", nil, allowed, "id ASC", 10, "", nil, &dest)
		if err != nil {
			t.Fatalf("searchWithRelevance() error = %v", err)
		}
		if dest != nil {
			t.Fatalf("dest = %v, want nil (untouched)", dest)
		}
	})
}

func TestSearchWithRelevanceInvalidField(t *testing.T) {
	db := setupSearchTestDB(t)
	allowed := map[string]bool{"name": true}

	var dest []searchResult
	err := searchWithRelevance(db, "search_items", "apple", []string{"name", "secret"}, allowed, "id ASC", 10, "", nil, &dest)
	if err == nil {
		t.Fatalf("searchWithRelevance() error = nil, want error")
	}
	wantErr := "invalid search field: secret"
	if err.Error() != wantErr {
		t.Fatalf("searchWithRelevance() error = %q, want %q", err.Error(), wantErr)
	}
}

func TestSearchWithRelevanceNoMatches(t *testing.T) {
	db := setupSearchTestDB(t)
	seedSearchItems(t, db, searchItem{Name: "banana"})
	allowed := map[string]bool{"name": true}

	var dest []searchResult
	err := searchWithRelevance(db, "search_items", "apple", []string{"name"}, allowed, "id ASC", 10, "", nil, &dest)
	if err != nil {
		t.Fatalf("searchWithRelevance() error = %v", err)
	}
	if len(dest) != 0 {
		t.Fatalf("dest = %v, want empty", dest)
	}
}

func TestSearchWithRelevanceRanksCloserMatchesHigher(t *testing.T) {
	db := setupSearchTestDB(t)
	seedSearchItems(t, db,
		searchItem{Name: "green apple"},
		searchItem{Name: "apple"},
		searchItem{Name: "apple pie"},
		searchItem{Name: "banana"},
	)
	allowed := map[string]bool{"name": true}

	var dest []searchResult
	err := searchWithRelevance(db, "search_items", "apple", []string{"name"}, allowed, "id ASC", 10, "", nil, &dest)
	if err != nil {
		t.Fatalf("searchWithRelevance() error = %v", err)
	}
	if len(dest) != 3 {
		t.Fatalf("len(dest) = %d, want 3; results = %+v", len(dest), dest)
	}

	// "apple" and "apple pie" both match at the start of the string (tied
	// relevance), and are ordered by the secondary "id ASC" column, which
	// puts "apple" (inserted second) before "apple pie" (inserted third).
	// "green apple" matches later in the string, so it ranks last.
	wantNames := []string{"apple", "apple pie", "green apple"}
	for i, want := range wantNames {
		if dest[i].Name != want {
			t.Fatalf("dest[%d].Name = %q, want %q (full results = %+v)", i, dest[i].Name, want, dest)
		}
	}
	for i := 0; i < len(dest)-1; i++ {
		if dest[i].Relevance < dest[i+1].Relevance {
			t.Fatalf("results not sorted by descending relevance: %+v", dest)
		}
	}
}

func TestSearchWithRelevanceIsCaseInsensitive(t *testing.T) {
	db := setupSearchTestDB(t)
	seedSearchItems(t, db, searchItem{Name: "Apple Pie"})
	allowed := map[string]bool{"name": true}

	var dest []searchResult
	err := searchWithRelevance(db, "search_items", "APPLE", []string{"name"}, allowed, "id ASC", 10, "", nil, &dest)
	if err != nil {
		t.Fatalf("searchWithRelevance() error = %v", err)
	}
	if len(dest) != 1 || dest[0].Name != "Apple Pie" {
		t.Fatalf("dest = %+v, want single match \"Apple Pie\"", dest)
	}
}

func TestSearchWithRelevanceMultipleFieldsUsesBestMatch(t *testing.T) {
	db := setupSearchTestDB(t)
	seedSearchItems(t, db,
		// "apple" matches in Name at the very start.
		searchItem{Name: "apple", Description: "a rare tropical fruit"},
		// "apple" only matches in Description, further from the start.
		searchItem{Name: "banana", Description: "tastes nothing like apple"},
	)
	allowed := map[string]bool{"name": true, "description": true}

	var dest []searchResult
	err := searchWithRelevance(db, "search_items", "apple", []string{"name", "description"}, allowed, "id ASC", 10, "", nil, &dest)
	if err != nil {
		t.Fatalf("searchWithRelevance() error = %v", err)
	}
	if len(dest) != 2 {
		t.Fatalf("len(dest) = %d, want 2; results = %+v", len(dest), dest)
	}
	if dest[0].Name != "apple" || dest[1].Name != "banana" {
		t.Fatalf("dest = %+v, want [apple, banana] order", dest)
	}
}

func TestSearchWithRelevanceExtraWhereFiltersResults(t *testing.T) {
	db := setupSearchTestDB(t)
	seedSearchItems(t, db,
		searchItem{Name: "apple", ProfileID: 1},
		searchItem{Name: "apple", ProfileID: 2},
	)
	allowed := map[string]bool{"name": true}

	var dest []searchResult
	err := searchWithRelevance(db, "search_items", "apple", []string{"name"}, allowed, "id ASC", 10, "profile_id = ?", []any{uint(2)}, &dest)
	if err != nil {
		t.Fatalf("searchWithRelevance() error = %v", err)
	}
	if len(dest) != 1 || dest[0].ProfileID != 2 {
		t.Fatalf("dest = %+v, want single result with ProfileID 2", dest)
	}
}

func TestSearchWithRelevanceRespectsLimit(t *testing.T) {
	db := setupSearchTestDB(t)
	seedSearchItems(t, db,
		searchItem{Name: "apple one"},
		searchItem{Name: "apple two"},
		searchItem{Name: "apple three"},
	)
	allowed := map[string]bool{"name": true}

	var dest []searchResult
	err := searchWithRelevance(db, "search_items", "apple", []string{"name"}, allowed, "id ASC", 2, "", nil, &dest)
	if err != nil {
		t.Fatalf("searchWithRelevance() error = %v", err)
	}
	if len(dest) != 2 {
		t.Fatalf("len(dest) = %d, want 2 (limit); results = %+v", len(dest), dest)
	}
}
