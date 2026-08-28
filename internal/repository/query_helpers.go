package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// resolveProfileID determines the profile ID to use, returning the provided ID if non-nil or the active
// profile ID otherwise. It queries the active profile state from the supplied profile repository if the input ID
// is nil. Returns the resolved profile ID or an error if the profile state retrieval fails.
func resolveProfileID(profileID *uint, profileRepo ProfileRepository) (uint, error) {
	if profileID != nil {
		return *profileID, nil
	}
	state, err := profileRepo.GetState()
	if err != nil {
		return 0, err
	}
	return state.ActiveCommandProfileID, nil
}

// resolveOrderClause validates and converts an order_by token into an SQL-compliant order clause (ASC/DESC).
func resolveOrderClause(token string, allowedColumns map[string]bool) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("empty order_by token")
	}
	desc := strings.HasPrefix(token, "-")
	column := token
	if desc {
		column = token[1:]
	}
	if !allowedColumns[column] {
		return "", fmt.Errorf("invalid order_by field: %s", token)
	}
	if desc {
		return column + " DESC", nil
	}
	return column + " ASC", nil
}

// resolveOrdering parses an orderBy string into SQL order clauses while validating against allowed columns.
// Returns a slice of valid SQL clauses or an error if any token is invalid.
func resolveOrdering(orderBy string, allowedColumns map[string]bool) ([]string, error) {
	tokens := splitCSV(orderBy)
	if len(tokens) == 0 {
		return nil, errors.New("empty order_by")
	}
	clauses := make([]string, 0, len(tokens))
	for _, token := range tokens {
		clause, err := resolveOrderClause(token, allowedColumns)
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
	}
	return clauses, nil
}

// splitCSV splits a comma-separated string into a slice of trimmed, non-empty substrings.
func splitCSV(items string) []string {
	parts := strings.Split(items, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// isUniqueConstraintViolation checks if the given error indicates a UNIQUE constraint violation
// for the specified table and column.
func isUniqueConstraintViolation(err error, table, column string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") &&
		strings.Contains(msg, table+"."+column)
}

// searchWithRelevanceInput defines the input structure for executing a full-text search with relevance ranking calculations.
type searchWithRelevanceInput struct {
	DB                   *gorm.DB
	Table                string
	Query                string
	Fields               []string
	AllowedColumns       map[string]bool
	SecondaryOrderColumn string
	Limit                int
	ExtraWhere           string
	ExtraArgs            []any
	Dest                 any
}

// searchWithRelevance performs a full-text search on the specified table, calculating relevance for query ranking.
func searchWithRelevance(input searchWithRelevanceInput) error {
	if input.Query == "" || len(input.Fields) == 0 {
		return nil
	}

	queryLower := strings.ToLower(input.Query)
	queryLen := len(queryLower)

	relevanceExprs := make([]string, 0, len(input.Fields))
	var relevanceArgs []any
	orClauses := make([]string, 0, len(input.Fields))
	var whereArgs []any

	for _, field := range input.Fields {
		if !input.AllowedColumns[field] {
			return fmt.Errorf("invalid search field: %s", field)
		}

		relevanceExprs = append(relevanceExprs, fmt.Sprintf(
			"((LENGTH(%s) - LENGTH(REPLACE(LOWER(%s), ?, ''))) / ? * 1000) - INSTR(LOWER(%s), ?)",
			field, field, field,
		))
		relevanceArgs = append(relevanceArgs, queryLower, queryLen, queryLower)

		orClauses = append(orClauses, fmt.Sprintf("LOWER(%s) LIKE ?", field))
		whereArgs = append(whereArgs, "%"+queryLower+"%")
	}

	relevanceSQL := relevanceExprs[0]
	if len(relevanceExprs) > 1 {
		relevanceSQL = "MIN(" + strings.Join(relevanceExprs, ", ") + ")"
	}

	where := strings.Join(orClauses, " OR ")
	args := append([]any{}, relevanceArgs...)
	args = append(args, whereArgs...)

	if input.ExtraWhere != "" {
		where = "(" + where + ") AND " + input.ExtraWhere
		args = append(args, input.ExtraArgs...)
	}

	sql := fmt.Sprintf(
		"SELECT %s.*, %s AS relevance FROM %s WHERE %s ORDER BY relevance DESC, %s LIMIT ?",
		input.Table, relevanceSQL, input.Table, where, input.SecondaryOrderColumn,
	)
	args = append(args, input.Limit)

	return input.DB.Raw(sql, args...).Scan(input.Dest).Error

}
