package anvil

type Rule struct{
	Name string
	Func func(value any) error
}

type Context interface {
	Field(name string, value any, rules ...Rule)

	Required() Rule
	Trim() Rule

	Min(min any) Rule
	Max(max any) Rule
	Range(min, max any) Rule
}
