package gen

import "github.com/nanoteck137/pyrin/tools/resolve"

type Generator interface {
	Name() string
	Generate(resolver *resolve.Resolver) error
}

type GeneratorStack struct {
	generators []Generator
}

func (stack *GeneratorStack) Generate(resolver *resolve.Resolver) error {
	for _, gen := range stack.generators {
		err := gen.Generate(resolver)
		if err != nil {
			return err
		}
	}

	return nil
}

func (stack *GeneratorStack) AddGenerator(generator Generator) {
	stack.generators = append(stack.generators, generator)
}
