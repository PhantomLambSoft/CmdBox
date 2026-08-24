package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

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

// searchWithRelevance performs a full-text search on the specified table, calculating relevance for query ranking.
// db: Database connection instance.
// table: Name of the target table for the search.
// query: The search query string.
// fields: List of fields to search within the table.
// allowedColumns: Map of fields allowed for searching to ensure security.
// secondaryOrderColumn: Column to use for secondary sorting after relevance ranking.
// limit: Maximum number of results to return.
// extraWhere: Additional SQL where conditions to apply to the query.
// extraArgs: Arguments for the extra where conditions.
// dest: Destination to store the query result.
// Returns an error if the query execution fails or invalid fields are provided.
func searchWithRelevance(
	db *gorm.DB,
	table string,
	query string,
	fields []string,
	allowedColumns map[string]bool,
	secondaryOrderColumn string,
	limit int,
	extraWhere string,
	extraArgs []any,
	dest any,
) error {
	if query == "" || len(fields) == 0 {
		return nil
	}

	queryLower := strings.ToLower(query)
	queryLen := len(queryLower)

	relevanceExprs := make([]string, 0, len(fields))
	var relevanceArgs []any
	orClauses := make([]string, 0, len(fields))
	var whereArgs []any

	for _, field := range fields {
		if !allowedColumns[field] {
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

	if extraWhere != "" {
		where = "(" + where + ") AND " + extraWhere
		args = append(args, extraArgs...)
	}

	sql := fmt.Sprintf(
		"SELECT %s.*, %s AS relevance FROM %s WHERE %s ORDER BY relevance DESC, %s LIMIT ?",
		table, relevanceSQL, table, where, secondaryOrderColumn,
	)
	args = append(args, limit)

	return db.Raw(sql, args...).Scan(dest).Error

}
