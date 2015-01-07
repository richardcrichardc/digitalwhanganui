package validate

import (
	"regexp"
)

func Required(field, key, label string, errors map[string]interface{}) {
	if field == "" {
		errors[key] = label + " cannot be blank."
	}
}

func Email(field, key, label string, errors map[string]interface{}) {
	matched, err := regexp.MatchString("^(?i:[A-Z0-9._%+-]+@[A-Z0-9.-]+)$", field)

	if err != nil {
		panic(err)
	}

	if !matched {
		errors[key] = label + " does not look like a valid email address. I am expecting to see something like: john.smith@example.com"
	}
}
