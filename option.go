package assign

// Option follows the option pattern for Assigner.
type Option func(*Assigner)

// WithTags appends tag keys to be used in struct assignment.
// This allows the destination Go struct to define tags to use
// while looking up fields by name in the source Source.
// The tag key is matched on the order given,
// starting with `assign` being the default tag key.
// The field name of the destination Go struct is used
// when tags are not matched by key for the field.
func WithTags(tags ...string) Option {
	return func(a *Assigner) {
		a.tags = append(a.tags, tags...)
	}
}

// WithoutCycle disables the cyclical path check.
// This is useful for Source types that do not have cycles.
// Source.Pointer is not called and need not be behave as expected for this case.
// Checking for cycles is enable by default.
// TODO: WithoutPointer? cyclical?
func WithoutCycle() Option {
	return func(a *Assigner) {
		a.cycle = false
	}
}
