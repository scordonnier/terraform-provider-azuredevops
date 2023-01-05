package utils

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

func StringNotEmptyValidator() validator.String {
	return stringvalidator.RegexMatches(regexp.MustCompile("^.*\\S.*$"), "must not be empty")
}

func UUIDStringValidator() validator.String {
	return stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"), "must be a valid UUID")
}
