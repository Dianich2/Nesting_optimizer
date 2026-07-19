package user

import "strings"

func normalizeIdentifier(
	field string,
) string {
	return strings.ToLower(
		strings.TrimSpace(field),
	)
}

func normalizeName(
	name string,
) string {
	return strings.TrimSpace(name)
}

func normalizeCreateUserInput(
	input CreateUserInput,
) CreateUserInput {
	input.Email = normalizeIdentifier(input.Email)
	input.Login = normalizeIdentifier(input.Login)
	input.FirstName = normalizeName(input.FirstName)
	input.LastName = normalizeName(input.LastName)
	return input
}

func normalizeLoginInput(
	input LoginInput,
) LoginInput {
	input.Identifier = normalizeIdentifier(input.Identifier)
	return input
}
