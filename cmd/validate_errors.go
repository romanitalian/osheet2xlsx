package cmd

import "errors"

// ErrValidateStructure signals structural validation failure for exit-code mapping.
var ErrValidateStructure = errors.New("validate: structure invalid")
