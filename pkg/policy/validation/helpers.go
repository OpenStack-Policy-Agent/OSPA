package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/policy"
)

// validateAllowedChecks validates that the check conditions only use allowed checks
// and that at least one allowed check is specified.
func validateAllowedChecks(check *policy.CheckConditions, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}

	allowedSet := make(map[string]bool, len(allowed))
	for _, name := range allowed {
		allowedSet[name] = true
	}

	// Get all set checks from the struct
	setChecks := getSetChecks(check)

	if len(setChecks) == 0 {
		return fmt.Errorf("check must specify at least one of: %s", strings.Join(allowed, ", "))
	}

	// Check if at least one allowed check is set
	hasAllowed := false
	for _, name := range setChecks {
		if allowedSet[name] {
			hasAllowed = true
			break
		}
	}
	if !hasAllowed {
		return fmt.Errorf("check must specify at least one of: %s", strings.Join(allowed, ", "))
	}

	// Find disallowed checks
	var disallowed []string
	for _, name := range setChecks {
		if !allowedSet[name] {
			disallowed = append(disallowed, name)
		}
	}
	if len(disallowed) > 0 {
		return fmt.Errorf("check specifies unsupported fields: %s", strings.Join(disallowed, ", "))
	}

	return nil
}

// getSetChecks returns the yaml tag names of all non-zero fields in CheckConditions.
// This dynamically discovers which checks are set based on the struct definition.
func getSetChecks(check *policy.CheckConditions) []string {
	var setChecks []string

	v := reflect.ValueOf(check).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get the yaml tag name
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		// Extract the field name from the tag (before any options like ",omitempty")
		tagName := strings.Split(yamlTag, ",")[0]
		if tagName == "" {
			continue
		}

		// Check if the field is set (non-zero value)
		if !isZeroValue(field) {
			setChecks = append(setChecks, tagName)
		}
	}

	return setChecks
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
