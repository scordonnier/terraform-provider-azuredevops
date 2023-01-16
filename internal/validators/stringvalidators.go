package validators

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"regexp"
)

func DateTimeValidator() validator.String {
	return stringvalidator.RegexMatches(regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$"), "must not be valid date (eg. 2000-12-25T00:00:00Z)")
}

func EnabledDisabledValidator() validator.String {
	return stringvalidator.OneOfCaseInsensitive("enabled", "disabled")
}

func PermissionsValidator() validator.String {
	return stringvalidator.OneOfCaseInsensitive("notset", "allow", "deny")
}

func StringNotEmptyValidator() validator.String {
	return stringvalidator.RegexMatches(regexp.MustCompile("^.*\\S.*$"), "must not be empty")
}

func UUIDStringValidator() validator.String {
	return stringvalidator.RegexMatches(regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"), "must be a valid UUID")
}
