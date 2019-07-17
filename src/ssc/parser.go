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

type Parser struct {
	defLexer      *Lexer
	configLexer   *Lexer
	sourceLexer   *Lexer
	varDeclareSet map[string]*VarDeclare
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
func ParseError(token *Token, m string) {
	log.Panicln(token.lineno, token.text, "ERR:", m)
}
/*
type DefRpc struct {
	isParamStream bool
	isRetStream   bool
	name          string
	param         string
	ret           string
}
type DefService struct {
	name    string
	rpcList []*DefRpc
}
type DefType struct {
	parentType string // "" or "xxx"
	def        string // "enum" or "message"
	name       string
	members    []*DefMember
}
type DefMember struct {
	typeName string
	name     string
	no       int
	tag      string // "" or "array"
}

type Parser struct {
	defLexer      *Lexer
	configLexer   *Lexer
	sourceLexer   *Lexer
	v             string
	pkg           string
	types         []*DefType
	services      []*DefService
	typeCheckList []*Token
}

func newParser(defLexer *Lexer, configLexer *Lexer) *Parser {
	p := &Parser{}
	p.defLexer = defLexer
	p.configLexer = configLexer
	p.types = make([]*DefType, 0)
	p.services = make([]*DefService, 0)
	p.typeCheckList = make([]*Token, 0)
	return p
}

func ParseError(token *Token, m string) {
	log.Panicln(token.lineno, token.text, "ERR:", m)
}

func (p *Parser) isBasicType(tokenType int) bool {
	//switch tokenType {
	//case TOKEN_KEYWORD_DOUBLE, TOKEN_KEYWORD_FLOAT, TOKEN_KEYWORD_INT32,
	//	TOKEN_KEYWORD_INT64, TOKEN_KEYWORD_UINT32, TOKEN_KEYWORD_UINT64,
	//	TOKEN_KEYWORD_SINT32, TOKEN_KEYWORD_SINT64, TOKEN_KEYWORD_FIXED32,
	//	TOKEN_KEYWORD_FIXED64, TOKEN_KEYWORD_SFIXED32, TOKEN_KEYWORD_SFIXED64,
	//	TOKEN_KEYWORD_BOOL, TOKEN_KEYWORD_STRING, TOKEN_KEYWORD_BYTES:
	//	return true
	//}
	return false
}

func (p *Parser) getDefType(name string) *DefType {
	for _, v := range p.types {
		if v.name == name {
			return v
		}
	}
	return nil
}

func (p *Parser) checkUnknownTypes() {
	for _, v := range p.typeCheckList {
		if p.isBasicType(v.tokenType) == false {
			if p.getDefType(v.text) == nil {
				ParseError(v, "unkonown type")
			}
		}
	}
}

func (p *Parser) printAll() {
	// print types
	for _, defType := range p.types {
		log.Printf("%s %s %s\n", defType.parentType, defType.def, defType.name)
		for _, defMember := range defType.members {
			log.Printf("	%s %s %s\n", defMember.tag, defMember.typeName, defMember.name)
		}
	}
	// print services
	for _, defService := range p.services {
		log.Printf("service %s\n", defService.name)
		for _, rpc := range defService.rpcList {
			if rpc.isParamStream && rpc.isRetStream {
				log.Printf("  %s (stream %s) (stream %s)\n", rpc.name, rpc.param, rpc.ret)
			} else if rpc.isParamStream {
				log.Printf("  %s (stream %s) (%s)\n", rpc.name, rpc.param, rpc.ret)
			} else if rpc.isRetStream {
				log.Printf("  %s (%s) (stream %s)\n", rpc.name, rpc.param, rpc.ret)
			} else {
				log.Printf("  %s (%s) (%s)\n", rpc.name, rpc.param, rpc.ret)
			}
		}
	}
}

func (p *Parser) checkToken(tokenType int) *Token {
	token := p.defLexer.takeToken()
	if token.tokenType != tokenType {
		ParseError(token, "invalid syntax")
	}
	//log.Println("checkToken", token.lineno, token.text)
	return token
}

func (p *Parser) parse(sourceLexer *Lexer) {
	p.sourceLexer = sourceLexer

	// 1. syntax token
	p.parse_syntax()

	// 2. package token
	p.parse_package()

	// 3. other tokens
	for {
		tokenType := p.configLexer.nextTokenType()
		if tokenType == -1 {
			break
		}
		switch tokenType {
		case TOKEN_KEYWORD_ENUM:
			p.parse_enum("")
		case TOKEN_KEYWORD_MESSAGE:
			p.parse_message("")
		case TOKEN_KEYWORD_SERVICE:
			p.parse_service()
		default:
			ParseError(p.configLexer.takeToken(), "proto error here")
		}
	}
	//!!!TODO: need support keyword import, or can not use checkUnknownTypes()
	//p.checkUnknownTypes()
	//p.printAll()
}

func (p *Parser) parse_syntax() {
	p.checkToken(TOKEN_KEYWORD_SYNTAX)   // syntax
	p.checkToken(TOKEN_ASSIGN)           // =
	p.checkToken(TOKEN_QUOTE)            // "
	syntax := p.checkToken(TOKEN_SYMBOL) // xxx
	p.checkToken(TOKEN_QUOTE)            // "
	p.checkToken(TOKEN_SEMICOLON)        // ;
	p.v = syntax.text
}

func (p *Parser) parse_import() {
	p.checkToken(TOKEN_KEYWORD_IMPORT)      // import
	p.checkToken(TOKEN_QUOTE)               // "
	importTxt := p.checkToken(TOKEN_SYMBOL) // xxx
	p.checkToken(TOKEN_QUOTE)               // "
	p.checkToken(TOKEN_SEMICOLON)           // ;
	p.v = importTxt.text
}

func (p *Parser) parse_package() {
	p.checkToken(TOKEN_KEYWORD_PACKAGE) // package
	pkg := p.checkToken(TOKEN_SYMBOL)   // xxx
	p.checkToken(TOKEN_SEMICOLON)       // ;
	p.pkg = pkg.text
}

func (p *Parser) parse_enum(parentType string) {
	//log.Println("parse_enum", parentType)
	p.checkToken(TOKEN_KEYWORD_ENUM)       // enum
	enumName := p.checkToken(TOKEN_SYMBOL) // xxx
	defType := &DefType{
		parentType: parentType,
		def:        "enum",
		name:       enumName.text,
		members:    make([]*DefMember, 0),
	}
	p.types = append(p.types, defType)
	p.checkToken(TOKEN_BRACE_LEFT) // {
	p.parse_enum_members(defType)
	p.checkToken(TOKEN_BRACE_RIGHT) // }
}

func (p *Parser) parse_enum_members(defType *DefType) {
	member := p.checkToken(TOKEN_SYMBOL) // xxx
	p.checkToken(TOKEN_ASSIGN)           // =
	no := p.checkToken(TOKEN_NUMBER)     // 1,2,3
	p.checkToken(TOKEN_SEMICOLON)        // ;
	num, _ := strconv.Atoi(no.text)
	defMember := &DefMember{
		typeName: "",
		name:     member.text,
		no:       num,
		tag:      "",
	}
	defType.members = append(defType.members, defMember)
	if p.configLexer.nextTokenType() == TOKEN_SYMBOL {
		p.parse_enum_members(defType)
	}
}

func (p *Parser) parse_message(parentType string) {
	//log.Println("parse_message", parentType)
	p.checkToken(TOKEN_KEYWORD_MESSAGE) // message
	var messageName *Token
	messageName = p.checkToken(TOKEN_SYMBOL) // xxx
	defType := &DefType{
		parentType: parentType,
		def:        "message",
		name:       messageName.text,
		members:    make([]*DefMember, 0),
	}
	p.types = append(p.types, defType)
	p.checkToken(TOKEN_BRACE_LEFT) // {
	p.parse_message_members(defType)
	p.checkToken(TOKEN_BRACE_RIGHT) // }
}

func (p *Parser) parse_message_members(defType *DefType) {
	nextTokenType := p.configLexer.nextTokenType()
	if nextTokenType == -1 {
		return
	}
	if p.v == "proto3" {
		switch nextTokenType {
		case TOKEN_KEYWORD_DOUBLE, TOKEN_KEYWORD_FLOAT, TOKEN_KEYWORD_INT32,
			TOKEN_KEYWORD_INT64, TOKEN_KEYWORD_UINT32, TOKEN_KEYWORD_UINT64,
			TOKEN_KEYWORD_SINT32, TOKEN_KEYWORD_SINT64, TOKEN_KEYWORD_FIXED32,
			TOKEN_KEYWORD_FIXED64, TOKEN_KEYWORD_SFIXED32, TOKEN_KEYWORD_SFIXED64,
			TOKEN_KEYWORD_BOOL, TOKEN_KEYWORD_STRING, TOKEN_KEYWORD_BYTES,
			TOKEN_SYMBOL, TOKEN_SYMBOL2, TOKEN_KEYWORD_REPEATED:
			tag := ""
			if nextTokenType == TOKEN_KEYWORD_REPEATED {
				tag = "array"
				p.checkToken(TOKEN_KEYWORD_REPEATED)
			}
			memberType := p.checkToken(p.configLexer.nextTokenType()) // xxx
			p.typeCheckList = append(p.typeCheckList, memberType)
			member := p.checkToken(TOKEN_SYMBOL) // xxx
			p.checkToken(TOKEN_ASSIGN)           // =
			no := p.checkToken(TOKEN_NUMBER)     // 1,2,3
			p.checkToken(TOKEN_SEMICOLON)        // ;
			num, _ := strconv.Atoi(no.text)
			defMember := &DefMember{
				typeName: memberType.text,
				name:     member.text,
				no:       num,
				tag:      tag,
			}
			defType.members = append(defType.members, defMember)

		case TOKEN_KEYWORD_ENUM:
			p.parse_enum(defType.name)
		case TOKEN_KEYWORD_MESSAGE:
			p.parse_message(defType.name)
		// function end
		default:
			return
		}

	} else if p.v == "proto2" {

	}

	p.parse_message_members(defType)
}

func (p *Parser) parse_service() {
	//log.Println("parse_service")
	p.checkToken(TOKEN_KEYWORD_SERVICE)       // service
	serviceName := p.checkToken(TOKEN_SYMBOL) // xxx
	service := &DefService{
		name:    serviceName.text,
		rpcList: make([]*DefRpc, 0),
	}
	p.services = append(p.services, service)
	p.checkToken(TOKEN_BRACE_LEFT) // {
	p.parse_service_rpcs(service)
	p.checkToken(TOKEN_BRACE_RIGHT) // }
}

func (p *Parser) parse_service_rpcs(service *DefService) {
	if p.configLexer.nextTokenType() != TOKEN_KEYWORD_RPC {
		return
	}
	p.checkToken(TOKEN_KEYWORD_RPC)       // rpc
	rpcName := p.checkToken(TOKEN_SYMBOL) //xxx
	p.checkToken(TOKEN_BRACKETS_LEFT)     // (

	isParamStream := false
	if p.configLexer.nextTokenType() == TOKEN_KEYWORD_STREAM {
		p.checkToken(TOKEN_KEYWORD_STREAM) // stream
		isParamStream = true
	}
	paramType := p.checkToken(TOKEN_SYMBOL) // xxx
	p.checkToken(TOKEN_BRACKETS_RIGHT)      // )

	p.checkToken(TOKEN_KEYWORD_RETURNS) // returns

	p.checkToken(TOKEN_BRACKETS_LEFT) // (
	isRetStream := false
	if p.configLexer.nextTokenType() == TOKEN_KEYWORD_STREAM {
		p.checkToken(TOKEN_KEYWORD_STREAM) // stream
		isRetStream = true
	}
	retType := p.checkToken(TOKEN_SYMBOL) // xxx
	p.checkToken(TOKEN_BRACKETS_RIGHT)    // )

	if p.configLexer.nextTokenType() == TOKEN_BRACE_LEFT {
		p.checkToken(TOKEN_BRACE_LEFT)
		p.checkToken(TOKEN_BRACE_RIGHT)
	}

	p.checkToken(TOKEN_SEMICOLON) // ;

	p.typeCheckList = append(p.typeCheckList, paramType)
	p.typeCheckList = append(p.typeCheckList, retType)
	service.rpcList = append(service.rpcList, &DefRpc{
		name:          rpcName.text,
		isParamStream: isParamStream,
		isRetStream:   isRetStream,
		param:         paramType.text,
		ret:           retType.text,
	})

	p.parse_service_rpcs(service)
}
*/