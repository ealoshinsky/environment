//go:build development
// +build development

package environment

import (
	"log"
)

const defaultEnvironmentFile = ".env"

func RegisterEnvironment[T any](instance *T) {
	if err := fillSpecification[T](instance, defaultEnvironmentFile); err != nil {
		log.Fatalf("%v", err)
	}
}
