package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"testing"
)

func TestParseProgram(t *testing.T) {
	input := `
		let x = 5;
		let y = 10;
		let foobar = 838383;
	`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program, 3)

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
	}
	for i, test := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, test.expectedIdentifier) {
			return
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}
	testIdentifier(t, stmt.Expression, input)
}

func TestBooleanExpression(t *testing.T) {
	input := []struct {
		input    string
		expected bool
	}{
		{"false", false},
		{"true", true},
	}
	for _, test := range input {
		l := lexer.New(test.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		checkProgram(t, program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
		}
		testBoolean(t, stmt.Expression, test.expected)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}
	ident, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("Expected Identifier, got %T", stmt.Expression)
	}
	if ident.Value != 5 {
		t.Errorf("Expected '%s', got '%d'", input, ident.Value)
	}
	if ident.TokenLiteral() != input {
		t.Errorf("Expected '%s', got '%s'", input, ident.TokenLiteral())
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world"`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got=%T", program.Statements[0])
	}
	str, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got=%T", stmt.Expression)
	}

	if str.Value != "hello world" {
		t.Fatalf("expected: %s, got: %s", "hello world", input)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		input        string
		operator     string
		integerValue interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true;", "!", true},
		{"-false;", "-", false},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		checkProgram(t, program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
		}
		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("Expected PrefixExpression, got %T", stmt.Expression)
		}
		if exp.Operator != test.operator {
			t.Errorf("Expected '%s', got '%s'", test.operator, exp.Operator)
		}

		if !testLiteralExpression(t, exp.Right, test.integerValue) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 < 5", 5, "<", 5},
		{"5 > 5", 5, ">", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		checkProgram(t, program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
		}
		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("Expected InfixExpression, got %T", stmt.Expression)
		}
		testInfixExpression(t, exp, test.leftValue, test.operator, test.rightValue)

	}
}

func checkProgram(t *testing.T, program *ast.Program, length ...int) {
	if program == nil {
		t.Fatalf("ParserProgram() returned nil")
	}
	count := 1
	if len(length) > 0 {
		count = length[0]
	}

	if len(program.Statements) != count {
		t.Fatalf("Expected %d statement(s), got=%d", count, len(program.Statements))
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a + b * c", "(a + (b * c))"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
		{"1 + (2 + 3) + 4", "((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2", "((5 + 5) * 2)"},
		{"2 / (5 + 5)", "(2 / (5 + 5))"},
		{"-(5 + 5)", "(-(5 + 5))"},
		{"!(true == true)", "(!(true == true))"},
		{"a + add(b * c) + d", "((a + add((b * c))) + d)"},
		{"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))", "add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))"},
		{"add(a + b + c * d / f + g)", "add((((a + b) + ((c * d) / f)) + g))"},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != test.expected {
			t.Errorf("Expected '%s', got '%s'", test.expected, actual)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}
	ie, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("Expected IfExpression, got %T", stmt.Expression)
	}
	if !testInfixExpression(t, ie.Condition, "x", "<", "y") {
		return
	}

	if len(ie.Consequence.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(ie.Consequence.Statements))
	}

	consequence, ok := ie.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", ie.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	if ie.Alternative != nil {
		t.Fatalf("Expected nil, got %T", ie.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}
	ie, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("Expected IfExpression, got %T", stmt.Expression)
	}
	if !testInfixExpression(t, ie.Condition, "x", "<", "y") {
		return
	}

	if len(ie.Consequence.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(ie.Consequence.Statements))
	}

	consequence, ok := ie.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", ie.Consequence.Statements[0])
	}

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	alternative, ok := ie.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", ie.Alternative.Statements[0])
	}
	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}
	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("Expected FunctionLiteral, got %T", stmt.Expression)
	}
	if len(function.Parameters) != 2 {
		t.Fatalf("Expected 2 params, got %d", len(function.Parameters))
	}
	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", function.Body.Statements[0])
	}
	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"fn(x, y) { x + y; }", []string{"x", "y"}},
		{"fn(x) { x + y; }", []string{"x"}},
		{"fn() { x + y; }", []string{}},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		checkProgram(t, program)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
		}
		function, ok := stmt.Expression.(*ast.FunctionLiteral)
		if !ok {
			t.Fatalf("Expected FunctionLiteral, got %T", stmt.Expression)
		}
		if len(function.Parameters) != len(test.expected) {
			t.Fatalf("Expected %d params, got %d", len(test.expected), len(function.Parameters))
		}
		for i, param := range test.expected {
			testLiteralExpression(t, function.Parameters[i], param)
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := `add(1, 2 * 3, 4 + 5)`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkProgram(t, program)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}
	call, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("Expected CallExpression, got %T", stmt.Expression)
	}
	if !testIdentifier(t, call.Function, "add") {
		return
	}
	if len(call.Arguments) != 3 {
		t.Fatalf("Expected 3 arguments, got %d", len(call.Arguments))
	}
	testLiteralExpression(t, call.Arguments[0], 1)
	testInfixExpression(t, call.Arguments[1], 2, "*", 3)
	testInfixExpression(t, call.Arguments[2], 4, "+", 5)
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := New(l)

		program := p.ParseProgram()
		checkParserErrors(t, p)

		checkProgram(t, program, 1)

		stmt, ok := program.Statements[0].(*ast.LetStatement)
		if !ok {
			t.Fatalf("Expected LetStatement, got %T", program.Statements[0])
		}
		if !testLetStatement(t, stmt, test.expectedIdentifier) {
			return
		}
		if !testLiteralExpression(t, stmt.Value, test.expectedValue) {
			return
		}
	}
}

func testIntegerLiteralExpression(test *testing.T, expr ast.Expression, value int64) bool {
	il, ok := expr.(*ast.IntegerLiteral)
	if !ok {
		test.Errorf("Expected IntegerLiteral, got %T", expr)
		return false
	}
	if il.Value != value {
		test.Errorf("Expected '%d', got '%d'", value, il.Value)
		return false
	}
	if il.TokenLiteral() != fmt.Sprintf("%d", value) {
		test.Errorf("Expected '%d', got '%s'", value, il.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(test *testing.T, expression ast.Expression, value string) bool {
	ident, ok := expression.(*ast.Identifier)
	if !ok {
		test.Errorf("Expected Identifier, got %T", expression)
		return false
	}
	if ident.Value != value {
		test.Errorf("Expected '%s', got '%s'", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != value {
		test.Errorf("Expected '%s', got '%s'", value, ident.TokenLiteral())
		return false
	}
	return true
}

func testBoolean(test *testing.T, expression ast.Expression, value bool) bool {
	ident, ok := expression.(*ast.Boolean)
	if !ok {
		test.Errorf("Expected Boolean, got %T", expression)
		return false
	}
	if ident.Value != value {
		test.Errorf("Expected '%t', got '%t'", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != fmt.Sprintf("%t", value) {
		test.Errorf("Expected '%t', got '%s'", value, ident.TokenLiteral())
		return false
	}
	return true
}

func testLiteralExpression(test *testing.T, expression ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteralExpression(test, expression, int64(v))
	case int64:
		return testIntegerLiteralExpression(test, expression, v)
	case string:
		return testIdentifier(test, expression, v)
	case bool:
		return testBoolean(test, expression, v)
	default:
		test.Errorf("Expected %T, got %T", expected, expression)
		return false
	}
}

func testInfixExpression(test *testing.T, expression ast.Expression, left interface{}, operator string, right interface{}) bool {
	infix, ok := expression.(*ast.InfixExpression)
	if !ok {
		test.Errorf("Expected InfixExpression, got %T", expression)
		return false
	}
	if !testLiteralExpression(test, infix.Left, left) {
		return false
	}
	if infix.Operator != operator {
		test.Errorf("Expected '%s', got '%s'", operator, infix.Operator)
		return false
	}
	if !testLiteralExpression(test, infix.Right, right) {
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	if len(errors) == 0 {
		return
	}

	for _, message := range errors {
		t.Errorf("parser error: %s", message)
	}
	t.FailNow()
}

func testLetStatement(test *testing.T, stmt ast.Statement, name string) bool {
	if stmt.TokenLiteral() != "let" {
		test.Errorf("Expected token 'let', got %s", stmt.TokenLiteral())
		return false
	}
	letStmt, ok := stmt.(*ast.LetStatement)
	if !ok {
		test.Errorf("Expected LetStatement, got %T", stmt)
		return false
	}
	if letStmt.Name.Value != name {
		test.Errorf("Expected name '%s', got '%s'", name, letStmt.Name.Value)
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		test.Errorf("Expected name '%s', got '%s'", name, letStmt.Name.TokenLiteral())
		return false
	}
	return true
}
