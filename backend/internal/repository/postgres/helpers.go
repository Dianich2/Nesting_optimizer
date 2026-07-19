package postgres

import (
	"errors"

	"github.com/lib/pq"
)

func uniqueViolationConstraint(err error) (string, bool) {
	var pqErr *pq.Error

	if !errors.As(err, &pqErr) {
		return "", false
	}

	if string(pqErr.Code) != postgresUniqueViolationCode {
		return "", false
	}

	return pqErr.Constraint, true
}
