package generator

import (
	nscgenerator "github.com/larkvincer/dsl-fsm/generator/nestedswitchcasegenerator"
	"github.com/larkvincer/dsl-fsm/optimizer"
)

type LanguageCodeGenerator interface {
	GetImplementer() nscgenerator.NSCNodeVisitor
	WriteFiles()
}

type CodeGenerator struct {
	optimizedStateMachine *optimizer.OptimizedStateMachine
	languageCodeGenerator LanguageCodeGenerator
}

func NewCodeGenerator(
	ost *optimizer.OptimizedStateMachine,
	languageCodeGenerator LanguageCodeGenerator,
) *CodeGenerator {
	return &CodeGenerator{
		optimizedStateMachine: ost,
		languageCodeGenerator: languageCodeGenerator,
	}
}

func (cg *CodeGenerator) Generate() {
	implementor := cg.languageCodeGenerator.GetImplementer()
	nscGenerator := nscgenerator.NSCGenerator{}
	nscGenerator.Generate(cg.optimizedStateMachine).Accept(implementor)
	cg.languageCodeGenerator.WriteFiles()
}
