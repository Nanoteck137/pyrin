package rules

import "github.com/nanoteck137/pyrin/tools/validate"

type RuleFunc func(value interface{}) error

type inlineRule struct {
	f RuleFunc
}

func (r *inlineRule) Validate(value interface{}) error {
	return r.f(value)
}

func By(f RuleFunc) validate.Rule {
	return &inlineRule{f: f}
}
