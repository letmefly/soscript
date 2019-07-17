package main
import("log")
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

	log.Println(varDeclare.name, varDeclare.currVal)
}

func (p *Parser) parseSourceCode(sourceLexer *Lexer) {
	p.sourceLexer = sourceLexer
	p.soscriptList = make([]*Soscript, 0)
	for p.sourceLexer.nextTokenType() != -1 {
		token := p.sourceLexer.takeToken()
		log.Println(token.lineno, token.tokenType, token.text)
		switch token.tokenType {
		case TAG_SOSCRIPT_START:
			p.parse_soscript(token)
		default:
			ParseError(token, "syntax error!")
		}
	}
}

func (p *Parser) parse_soscript(token *Token) {
	soscript :=  &Soscript{startLineno: token.lineno}
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

}

func (p *Parser) parse_soscript_if(soscript *Soscript, token *Token) {

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
	token := p.configLexer.takeToken()
	if token.tokenType != tokenType {
		ParseError(token, "invalid syntax")
	}
	//log.Println("checkToken", token.lineno, token.text)
	return token
}
func ParseError(token *Token, m string) {
	log.Panicln(token.lineno, token.text, "ERR:", m)
}
