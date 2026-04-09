package helpers

import "github.com/hashicorp/terraform-plugin-framework/types"

// StringPtr returns a pointer to the string value, or nil if null/unknown.
func StringPtr(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// IntPtr returns a pointer to the int value, or nil if null/unknown.
func IntPtr(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

// BoolPtr returns a pointer to the bool value, or nil if null/unknown.
func BoolPtr(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}
