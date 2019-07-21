package main
import(
	"fmt"
	"log")
/*
BNF Design:
<variable_declare> ::= <identifier>: {<variable_val>}

<variable_val> ::= <variable_val>, <const_val>      |
					<const_val>

<const_val> ::= <number>      |
				<string>

<identifier> ::= "\w+"

<variable_assign> ::= <identifier> = <const_val>

<if_expr> ::= if(<logic_calc_expr>) <print_expr>     | 
			  if(<logic_calc_expr>) <variable_assign> 

<logic_calc_expr> ::= <logic_calc_expr> || <logic_term>    |
					<logic_term>

<logic_term> ::= <identifier> && <logic_term>     |
					<identifier>

<print_expr> ::= print(<code> <code_expr> </code>)

<code_expr> ::= <string>

<tag> ::= "<soscript>" | "</soscript>" | "<default>" | "</default>" | "<line>" | "</line>" | "<code>" | "</code>" | "<var>" | "</var>" 
*/

type VarDeclare struct {
	name string
	varType string
	valList []string
	scope string
	currVal string
}

type Soscript struct {
	startLineno int
	endLineno int
	varDeclareSet map[string]*VarDeclare
}

type Parser struct {
	defLexer      *Lexer
	configLexer   *Lexer
	sourceLexer   *Lexer
	varDeclareSet map[string]*VarDeclare
	soscriptList []*Soscript
}

func newParser(defLexer *Lexer, configLexer *Lexer) *Parser {
	p := &Parser{
		varDeclareSet: make(map[string]*VarDeclare, 0),
		defLexer: defLexer,
		configLexer: configLexer,
	}
	p.init()
	return p
}

func (p *Parser) init() {
	p.parseDef()
	p.parseConfig()
}

func (p *Parser) parseDef() {
	for p.defLexer.nextTokenType() != -1 {
		token := p.defLexer.takeToken()
		switch token.tokenType {
		case TOKEN_SYMBOL:
			if p.defLexer.nextTokenType() == TOKEN_COLON {
				p.parse_var_declare(token)
			}
		default:
			ParseError(token, "syntax error!")
		}
	}
}

func (p *Parser) parse_var_declare(token *Token)  {
	_, ok := p.varDeclareSet[token.text]
	if ok {
		ParseError(token, "this variable has been decleared!");
	}
	varName := token.text
	p.varDeclareSet[varName] = &VarDeclare{name: varName, varType: "", valList: make([]string, 0), scope: "GLOBAL"}
	p.checkDefToken(TOKEN_COLON)
	p.checkDefToken(TOKEN_BRACE_LEFT)
	p.parse_var_declare_val(varName)
	p.checkDefToken(TOKEN_BRACE_RIGHT)
}

func (p *Parser) parse_var_declare_val(varName string) {
	if p.defLexer.nextTokenType() != TOKEN_NUMBER && p.defLexer.nextTokenType() != TOKEN_STRING {
		return
	}
	token := p.defLexer.takeToken()
	if token.tokenType == TOKEN_NUMBER {
		p.addGlobalVarVal(varName, "NUMBER", token)
	} else if token.tokenType == TOKEN_STRING {
		p.addGlobalVarVal(varName, "STRING", token)
	}
	if p.defLexer.nextTokenType() == TOKEN_COMMA {
		p.checkDefToken(TOKEN_COMMA)
		p.parse_var_declare_val(varName)
	}
}

func (p *Parser) addGlobalVarVal(varName string, varType string, token *Token) {
	varDeclare := p.varDeclareSet[varName]
	if varDeclare.varType != "" && varDeclare.varType != varType {
		ParseError(token, "var type not the same!")
	}
	varDeclare.varType = varType
	
	varDeclare.valList = append(varDeclare.valList, token.text)
	//log.Println("Global varible ", varName, varType, token.text)
}

func (p *Parser) parseConfig() {
	for p.configLexer.nextTokenType() != -1 {
		token := p.configLexer.takeToken()
		switch token.tokenType {
		case TOKEN_SYMBOL:
			if p.configLexer.nextTokenType() == TOKEN_ASSIGN {
				p.parse_assign(token)
			}
		default:
			ParseError(token, "syntax error!")
		}
	}
}

func (p *Parser) parse_assign(token *Token) {
	varName := token.text
	varDeclare, ok := p.varDeclareSet[varName]
	if !ok {
		ParseError(token, "this var not declared!")
	}
	p.checkConfigToken(TOKEN_ASSIGN)
	var valToken *Token
	if varDeclare.varType == "NUMBER" {
		valToken = p.checkConfigToken(TOKEN_NUMBER)
	} else if varDeclare.varType == "STRING" {
		valToken = p.checkConfigToken(TOKEN_STRING)
	}
	isDeclare := false
	for _, v := range varDeclare.valList {
		if v == valToken.text {
			isDeclare = true
			break
		}
	}
	if isDeclare == false {
		ParseError(valToken, "this var value not declared!")
	}
	varDeclare.currVal = valToken.text

	//log.Println(varDeclare.name, varDeclare.currVal)
}

func (p *Parser) parseSourceCode(sourceLexer *Lexer) {
	p.sourceLexer = sourceLexer
	p.soscriptList = make([]*Soscript, 0)
	for p.sourceLexer.nextTokenType() != -1 {
		token := p.sourceLexer.takeToken()
		//log.Println(token.lineno, token.tokenType, token.text)
		switch token.tokenType {
		case TAG_SOSCRIPT_START:
			p.parse_soscript(token)
		default:
			ParseError(token, "syntax error!")
		}
	}
}

func (p *Parser) parse_soscript(token *Token) {
	soscript :=  &Soscript{startLineno: token.lineno, varDeclareSet: make(map[string]*VarDeclare, 0)}
	p.soscriptList = append(p.soscriptList, soscript)
	p.checkSourceToken(TAG_DEFAULT_START)
	p.checkSourceToken(TAG_DEFAULT_END)
	for p.sourceLexer.nextTokenType() == TAG_LINE_START {
		p.checkSourceToken(TAG_LINE_START)
		p.parse_soscript_line(soscript)
		p.checkSourceToken(TAG_LINE_END)
	}
	p.checkSourceToken(TAG_SOSCRIPT_END)
}

func (p *Parser) parse_soscript_line(soscript *Soscript) {
	for p.sourceLexer.nextTokenType() != -1 {
		token := p.sourceLexer.takeToken()
		switch token.tokenType {
		case TOKEN_SYMBOL:
			if p.sourceLexer.nextTokenType() == TOKEN_ASSIGN {
				p.parse_soscript_assign(soscript, token)
			}
		case TOKEN_KEYWORD_IF:
			p.parse_soscript_if(soscript, token)
		default:
			ParseError(token, "syntax error!")
		}
	}
}

func (p *Parser) parse_soscript_assign(soscript *Soscript, token *Token) {
	varName := token.text
	p.checkSourceToken(TOKEN_ASSIGN)
	varVal := fmt.Sprintf("%s", p.parse_logic_expr(soscript))
	varDeclare := &VarDeclare{name:varName, varType:"BOOL", currVal:varVal}
	soscript.varDeclareSet[varName] = varDeclare
}

func (p *Parser) parse_soscript_if(soscript *Soscript, token *Token) {
	p.checkSourceToken(TOKEN_BRACKETS_LEFT)
	p.parse_logic_expr(soscript)
	p.checkSourceToken(TOKEN_BRACKETS_RIGHT)
	p.checkSourceToken(TOKEN_KEYWORD_PRINT)
	p.checkSourceToken(TOKEN_BRACKETS_LEFT)
	p.checkSourceToken(TAG_CODE_START)
	p.parse_code_expr(soscript)
	p.checkSourceToken(TAG_CODE_END)
	p.checkSourceToken(TOKEN_BRACKETS_RIGHT)
}

func (p *Parser) parse_logic_expr(soscript *Soscript) bool {
	val := false
	for p.sourceLexer.nextTokenType() != -1 {
		switch p.sourceLexer.nextTokenType() {
		case TOKEN_KEYWORD_OR:
			val = p.parse_logic_expr_or(soscript, p.sourceLexer.takeToken())
		case TOKEN_KEYWORD_AND:
			val = p.parse_logic_expr_and(soscript, p.sourceLexer.takeToken())
		case TOKEN_BRACKETS_LEFT:
			p.checkSourceToken(TOKEN_BRACKETS_LEFT)
			for p.sourceLexer.nextTokenType() != TOKEN_BRACKETS_RIGHT {
				val = p.parse_logic_expr(soscript)
			}
			p.checkSourceToken(TOKEN_BRACKETS_RIGHT)
		case TOKEN_SYMBOL:
			token := p.sourceLexer.takeToken()
			switch p.sourceLexer.nextTokenType() {
			case TOKEN_KEYWORD_OR:
				val = p.parse_logic_expr_or(soscript, token)
			case TOKEN_KEYWORD_AND:
				val = p.parse_logic_expr_and(soscript, token)
			case TOKEN_KEYWORD_NOT:
				val = p.parse_logic_expr_not(soscript, token)
			case TOKEN_EQUAL:
				val = p.parse_logic_expr_equel(soscript, token)
			case TOKEN_GREAT:
			case TOKEN_LESS:
			case TOKEN_GREAT_EQUAL:
			case TOKEN_LESS_EQUAL:
			default:
				val = p.parse_code_expr_symbol(soscript, token)
				break
			}
		case TOKEN_BRACKETS_RIGHT: break
		}

	}

	return val
}

func (p *Parser) parse_logic_expr_and(soscript *Soscript, token *Token) bool {
	val := p.parse_code_expr_symbol(soscript, token)
	p.checkSourceToken(TOKEN_KEYWORD_AND)
	return val && p.parse_logic_expr(soscript)
}

func (p *Parser) parse_logic_expr_or(soscript *Soscript, token *Token) bool {
	val := p.parse_code_expr_symbol(soscript, token)
	p.checkSourceToken(TOKEN_KEYWORD_OR)
	return val || p.parse_logic_expr(soscript)
}

func (p *Parser) parse_logic_expr_not(soscript *Soscript, token *Token) bool {
	p.checkSourceToken(TOKEN_KEYWORD_NOT)
	val := p.parse_code_expr_symbol(soscript, token)
	return !val
}

func (p *Parser) parse_logic_expr_equel(soscript *Soscript, token *Token) bool {
	varDef := p.checkVar(soscript, token.text)
	val := false
	p.checkSourceToken(TOKEN_EQUAL)
	rightToken := p.sourceLexer.takeToken()
	switch rightToken.tokenType {
	case TOKEN_NUMBER:
		val = varDef.currVal == rightToken.text
	case TOKEN_STRING:
		val = varDef.currVal == rightToken.text
	case TOKEN_SYMBOL:
		val = p.parse_code_expr_symbol(soscript, rightToken)
	}
	return !val
}

func (p *Parser) parse_code_expr_symbol(soscript *Soscript, token *Token) bool {
	val := false
	if "TRUE" == soscript.varDeclareSet[token.text].currVal {
		val = true
	} else if "FALSE" == soscript.varDeclareSet[token.text].currVal {
		val = false
	} else {
		ParseError(token, "value type error!")
	}
	return val
}

func (p *Parser) parse_code_expr(soscript *Soscript) {

}

func (p *Parser) checkVar(sososcript *Soscript, varName string) *VarDeclare {
	valDef, ok := p.varDeclareSet[varName]
	if ok {
		return valDef
	}
	valDef, ok = sososcript.varDeclareSet[varName]
	if ok {
		return valDef
	}
	log.Panicln("ERR: no var defined in config file for "+ varName)
	return nil
}

func (p *Parser) checkDefToken(tokenType int) *Token {
	token := p.defLexer.takeToken()
	if token.tokenType != tokenType {
		ParseError(token, "invalid syntax")
	}
	//log.Println("checkToken", token.lineno, token.text)
	return token
}
func (p *Parser) checkConfigToken(tokenType int) *Token {
	token := p.configLexer.takeToken()
	if token.tokenType != tokenType {
		ParseError(token, "invalid syntax")
	}
	//log.Println("checkToken", token.lineno, token.text)
	return token
}
func (p *Parser) checkSourceToken(tokenType int) *Token {
	token := p.sourceLexer.takeToken()
	if token.tokenType != tokenType {
		ParseError(token, "invalid syntax")
	}
	//log.Println("checkToken", token.lineno, token.text)
	return token
}
func ParseError(token *Token, m string) {
	log.Panicln(token.lineno, token.text, "ERR:", m)
}
