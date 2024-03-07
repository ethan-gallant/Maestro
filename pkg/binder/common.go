package binder

import (
	"errors"
)

type ContextKey string

var ErrStateNotFound = errors.New("state not found in context")
var ErrStateMismatch = errors.New("state type mismatch")
var ErrContextExists = errors.New("state already exists in context")
