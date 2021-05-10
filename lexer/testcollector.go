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

func (collector *TestCollector) openBrace(lineNumber int, position int) {
	collector.addToken("openBrace")
}

func (collector *TestCollector) closeBrace(lineNumber int, position int) {
	collector.addToken("closeBrace")
}

func (collector *TestCollector) openParen(lineNumber int, position int) {
	collector.addToken("openParen")
}

func (collector *TestCollector) closeParen(lineNumber int, position int) {
	collector.addToken("closeParen")
}

func (collector *TestCollector) openAngle(lineNumber int, position int) {
	collector.addToken("openAngle")
}

func (collector *TestCollector) closeAngle(lineNumber int, position int) {
	collector.addToken("closeAngle")
}

func (collector *TestCollector) star(lineNumber int, position int) {
	collector.addToken("star")
}

func (collector *TestCollector) colon(lineNumber int, position int) {
	collector.addToken("colon")
}

func (collector *TestCollector) name(name string, lineNumber int, position int) {
	collector.addToken("#" + name + "#")
}

func (collector *TestCollector) error(lineNumber int, position int) {
	collector.addToken("error")
}
