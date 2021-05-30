package generator

import (
	"fmt"

	"github.com/larkvincer/dsl-fsm/generator/implementors"
	nscgenerator "github.com/larkvincer/dsl-fsm/generator/nestedswitchcasegenerator"
)

type JavaCodeGenerator struct {
	javaNestedSwitchCaseImplementor *implementors.JavaNestedSwitchCaseImplementor
}

func NewJavaCodeGenerator(
	javaNestedSwitchCaseImplementor *implementors.JavaNestedSwitchCaseImplementor,
) *JavaCodeGenerator {
	return &JavaCodeGenerator{
		javaNestedSwitchCaseImplementor: javaNestedSwitchCaseImplementor,
	}
}

func (javaGenerator *JavaCodeGenerator) GetImplementer() nscgenerator.NSCNodeVisitor {
	return javaGenerator.javaNestedSwitchCaseImplementor
}

func (javaGenerator *JavaCodeGenerator) WriteFiles() {
	fmt.Println("Generated code:")
	fmt.Print(javaGenerator.javaNestedSwitchCaseImplementor.Output)
}
