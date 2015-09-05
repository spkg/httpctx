// Package env is used to determine whether the current program
// is running in a development, test or production environment.
package env

import (
	"regexp"
	"strings"

	"sp.com.au/exp/errs"
)

var (
	ErrNotTest     = errs.Forbidden("this operation can only be performed in a test environment")
	ErrInvalidName = errs.Forbidden("invalid environment name")
)

var (
	envname       string = DevelopmentPrefix
	isProduction  bool   = false
	isTest        bool   = false
	isDevelopment bool   = false
)

var (
	// Environment names must match this pattern
	envPattern = regexp.MustCompile(`^[a-z][a-z0-9]+$`)
)

// Environment name prefixes. An environment name must start
// with one of the following strings.
const (
	ProductionPrefix  = "production"
	TestPrefix        = "test"
	DevelopmentPrefix = "development"
)

var (
	callbacks []func() error
)

// FormatName is a function that will return an environment-specific name
// given the environment name, the separator character, and the base name
// of the object.
var FormatName func(envname, separator, name string) string

func init() {
	if FormatName == nil {
		FormatName = func(envname, separator, name string) string {
			return name + separator + envname
		}
	}
}

// Env returns the environment name.
func Name() string {
	return envname
}

// SetEnv sets the environment name. Returns
// an error if the name is not a valid environment name.
func SetName(name string) error {
	if !envPattern.MatchString(name) {
		return ErrInvalidName
	}

	prevEnvname := envname
	prevIsProduction := isProduction
	prevIsTest := isTest
	prevIsDevelopment := isDevelopment

	if strings.HasPrefix(name, ProductionPrefix) {
		envname = name
		isProduction = true
		isTest = false
		isDevelopment = false
	} else if strings.HasPrefix(name, TestPrefix) {
		envname = name
		isProduction = false
		isTest = true
		isDevelopment = false
	} else if strings.HasPrefix(name, DevelopmentPrefix) {
		envname = name
		isProduction = false
		isTest = false
		isDevelopment = true
	} else {
		return ErrInvalidName
	}

	for _, f := range callbacks {
		if err := f(); err != nil {
			envname = prevEnvname
			isProduction = prevIsProduction
			isTest = prevIsTest
			isDevelopment = prevIsDevelopment
			return err
		}
	}

	return nil
}

// MustSetEnv sets the environment name. If name is not a valid
// environment name, then the program panics.
func MustSetName(name string) {
	err := SetName(name)
	if err != nil {
		panic(err)
	}
}

// IsProduction returns true if the current environment is production.
func IsProduction() bool {
	return isProduction
}

// IsTest returns true if the current environment is test.
func IsTest() bool {
	return isTest
}

// CheckTest will return an error if the environment is not a test environment.
func CheckTest() error {
	if !isTest {
		return ErrNotTest
	}
	return nil
}

// IsDevelopment returns true if the current environment is development.
func IsDevelopment() bool {
	return isDevelopment
}

// NameFor is used to create environment-specific names for tables, queues
// and other schema objects. For example if the
// environment name is "test", then NameFor("tablename") returns
// "tablename_test".
func NameFor(name string) string {
	var separator string
	for _, s := range []string{"_", "-", "."} {
		if strings.Contains(name, s) {
			separator = s
			break
		}
	}
	if separator == "" {
		separator = "_"
	}

	return FormatName(envname, separator, name)
}

// WhenChanged allows a package to add a callback to be called
// whenever the environment changes.
func WhenChanged(callback func() error) {
	callbacks = append(callbacks, callback)
}
