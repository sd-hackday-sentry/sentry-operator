package controller

import (
	"github.com/sd-hackday-sentry/sentry-operator/pkg/controller/sentry"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, sentry.Add)
}
