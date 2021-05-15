package lexer

type TestCollector struct {
	tokens     string
	firstToken bool
}

func NewTestCollector() *TestCollector {
	return &TestCollector{firstToken: true}
}

func (collector *TestCollector) addToken(token string) {
	if !collector.firstToken {
		collector.tokens += ","
	}
	collector.tokens += token
	collector.firstToken = false
}

func (collector *TestCollector) OpenBrace(lineNumber int, position int) {
	collector.addToken("openBrace")
}

func (collector *TestCollector) CloseBrace(lineNumber int, position int) {
	collector.addToken("closeBrace")
}

func (collector *TestCollector) OpenParen(lineNumber int, position int) {
	collector.addToken("openParen")
}

func (collector *TestCollector) CloseParen(lineNumber int, position int) {
	collector.addToken("closeParen")
}

func (collector *TestCollector) OpenAngle(lineNumber int, position int) {
	collector.addToken("openAngle")
}

func (collector *TestCollector) CloseAngle(lineNumber int, position int) {
	collector.addToken("closeAngle")
}

func (collector *TestCollector) Star(lineNumber int, position int) {
	collector.addToken("star")
}

func (collector *TestCollector) Colon(lineNumber int, position int) {
	collector.addToken("colon")
}

func (collector *TestCollector) Name(name string, lineNumber int, position int) {
	collector.addToken("#" + name + "#")
}

func (collector *TestCollector) Error(lineNumber int, position int) {
	collector.addToken("error")
}
