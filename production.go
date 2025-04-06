//go:build !development
// +build !development

package environment

import "log"

func RegisterEnvironment[T any](instance *T) {
	if err := fillSpecification[T](instance); err != nil {
		log.Fatalf("%v", err)
	}
}
