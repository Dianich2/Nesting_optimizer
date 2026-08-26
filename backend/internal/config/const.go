package config

import "regexp"

const (
	MaxNameLen        = 150
	MaxDescriptionLen = 2000
	MaxUserNamesLen   = 50
	MaxEmailLen       = 254
	MaxLoginLen       = 100
	MinPasswordLen    = 8
	MaxPasswordLen    = 72
	MinPageSize       = 1
	MaxPageSize       = 100
	MinPage           = 1
	DefaultPage       = "1"
	DefaultPageSize   = "20"
)

var (
	RegexValidateEmail = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	RegexValidateLogin = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?$`)
	RegexValidateName  = regexp.MustCompile(`^\p{L}+([-'’ ]\p{L}+)*$`)
)
