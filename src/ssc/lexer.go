package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	TOKEN_MIN            = iota
	TOKEN_COMMENT         // //
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_EQUAL           // ==
	TOKEN_GREAT_EQUAL     // >=
	TOKEN_LESS_EQUAL      // <=
	TOKEN_GREAT           // >
	TOKEN_LESS			  // <
	TOKEN_ASSIGN          // =
	TOKEN_COMMA           // ,
	TOKEN_COLON           // :
	TOKEN_BRACE_LEFT      // {
	TOKEN_BRACE_RIGHT     // }
	TOKEN_BRACKETS_LEFT   // (
	TOKEN_BRACKETS_RIGHT  // )
	TOKEN_QUOTE           // "
	TOKEN_KEYWORD_IF      // if
	TOKEN_KEYWORD_PRINT   // print
	TOKEN_KEYWORD_AND     // &&
	TOKEN_KEYWORD_OR      // ||
	TOKEN_KEYWORD_NOT	  // !

	TOKEN_SYMBOL

	//TOKEN_DEFAULT_CODE

	TAG_SOSCRIPT_START
	TAG_SOSCRIPT_END

	TAG_DEFAULT_START
	TAG_DEFAULT_END

	TAG_LINE_START
	TAG_LINE_END

	TAG_CODE_START
	TAG_CODE_END

	TAG_VAR_START
	TAG_VAR_END

	TOKEN_MAX
)

var token_rules = map[int]string{
	TOKEN_COMMENT:        `\s*//\s*`,
	TOKEN_EQUAL:          `\s*==\s*`,
	TOKEN_ASSIGN:         `\s*=\s*`,
	TOKEN_COMMA:          `\s*\,\s*`,
	TOKEN_COLON:          `\s*\:\s*`,
	TOKEN_BRACE_LEFT:     `\s*{\s*`,
	TOKEN_BRACE_RIGHT:    `\s*}\s*`,
	TOKEN_BRACKETS_LEFT:  `\s*\(\s*`,
	TOKEN_BRACKETS_RIGHT: `\s*\)\s*`,
	//TOKEN_QUOTE:            `\s*"\s*`,
	TOKEN_KEYWORD_IF:    `\s*if\s*`,
	TOKEN_KEYWORD_PRINT: `\s*print\s*`,
	TOKEN_KEYWORD_AND:   `\s*&&\s*`,
	TOKEN_KEYWORD_OR:    `\s*\|\|\s*`,
	TOKEN_STRING:        `\s*"[^"]+"\s*`,
	TOKEN_SYMBOL:        `\s*[\w]+\s*`,
	TOKEN_NUMBER:        `\s*[\d]+\s*`,


	TAG_SOSCRIPT_START: `<soscript>`,
	TAG_SOSCRIPT_END: `</soscript>`,

	TAG_DEFAULT_START: `<default>`,
	TAG_DEFAULT_END: `</default>`,

	TAG_LINE_START: `<line>`,
	TAG_LINE_END: `</line>`,

	TAG_CODE_START: `<code>`,
	TAG_CODE_END: `</code>`,

	TAG_VAR_START:`<var>`,
	TAG_VAR_END: `</var>`,
}

type Token struct {
	lineno    int
	tokenType int
	text      string
}

type Lexer struct {
	fileType     string
	lines        []string
	rules        map[int]*regexp.Regexp
	tokens       []*Token
	currTokenIdx int

	in_soscript bool
	in_default bool
	//in_line bool
	//in_code bool
	//in_var bool
}

func newLexer(fileType string, reader io.Reader) *Lexer {
	lexer := &Lexer{
		in_soscript: false,
		in_default: false,
		//in_line: false,
		//in_code: false,
		//in_var: false,
	}
	lexer.init(fileType, reader)
	return lexer
}

func (lexer *Lexer) init(fileType string, reader io.Reader) {
	lexer.fileType = fileType
	lexer.currTokenIdx = 0
	lexer.rules = map[int]*regexp.Regexp{}
	for k, v := range token_rules {
		reg := regexp.MustCompile(v)
		lexer.rules[k] = reg
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		lexer.lines = append(lexer.lines, line)
	}
	for k, v := range lexer.lines {
		lexer.parseLine(k+1, v)
	}

	//for _, v := range lexer.tokens {
	//	log.Println("line ", v.lineno, v.text)
	//}
}

func (lexer *Lexer) parseLine(lineno int, lineText string) {
	line := lineText
	switch lexer.fileType {
	case "ss":
		lexer.start_ss(lineno, line)
	case "not_ss":
		lexer.start_not_ss(lineno, line)
	}
}

func (lexer *Lexer) start_ss(lineno int, line string) {
	for len(line) > 0 {
		isMatch := false
		for tokenType := TOKEN_MIN + 1; tokenType < TOKEN_MAX; tokenType++ {
			reg := lexer.rules[tokenType]
			if reg == nil {
				continue
			}
			ret := reg.FindStringIndex(line)
			if len(ret) == 2 && ret[0] == 0 {
				// do not process comment words
				if tokenType == TOKEN_COMMENT {
					return
				}
				text := strings.TrimSpace(line[ret[0]:ret[1]])
				lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: tokenType, text: text})
				line = line[ret[1]:]
				isMatch = true
				break
			}
		}
		if isMatch == false {
			fmt.Errorf("There is some error in proto file with line %v: %v", lineno, line)
			break
		}
	}
}

func (lexer *Lexer) start_not_ss(lineno int, line string) {
	// check: <soscript>
	if lexer.in_soscript == false {
		if lexer.rules[TAG_SOSCRIPT_START].FindStringIndex(line) != nil {
			lexer.in_soscript = true
			lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_SOSCRIPT_START, text: "<soscript>"})
			return
		}
	} else {
		lexer.do_in_soscript(lineno, line)
	}
}

func (lexer *Lexer) do_in_soscript(lineno int, line string) {
	// check: <default>
	if lexer.in_default == false {
		if lexer.rules[TAG_DEFAULT_START].FindStringIndex(line) != nil {
			lexer.in_default = true
			lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_DEFAULT_START, text: "<default>"})
			return
		}
	} else {
		lexer.do_in_default(lineno, line)
		return
	}

	// check: <line>
	ret := lexer.rules[TAG_LINE_START].FindStringIndex(line)
	if ret != nil {
		line := strings.TrimSpace(line[ret[0]:])
		//lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_LINE_START, text: "<line>"})
		lexer.do_in_line(lineno, line)
		return
	}

	// check: </soscript>
	if lexer.rules[TAG_SOSCRIPT_END].FindStringIndex(line) != nil {
		lexer.in_soscript = false
		lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_SOSCRIPT_END, text: "</soscript>"})
		return
	}
}

func (lexer *Lexer) do_in_default(lineno int, line string) {
	// check: </default>
	if lexer.rules[TAG_DEFAULT_END].FindStringIndex(line) != nil {
		lexer.in_default = false
		lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_DEFAULT_END, text: "</default>"})
		return
	}
}

func (lexer *Lexer) do_in_line(lineno int, line string) {
	//fmt.Println(lineno, line)
	for len(line) > 0 {
		isMatch := false
		for tokenType := TOKEN_MIN + 1; tokenType < TOKEN_MAX; tokenType++ {
			reg := lexer.rules[tokenType]
			if reg == nil {
				continue
			}
			ret := reg.FindStringIndex(line)
			if len(ret) == 2 && ret[0] == 0 {
				text := strings.TrimSpace(line[ret[0]:ret[1]])
				lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: tokenType, text: text})
				line = line[ret[1]:]
				isMatch = true
				// if current token is <code>
				if tokenType == TAG_CODE_START {
					// find </code>
					ret := lexer.rules[TAG_CODE_END].FindStringIndex(line)
					if ret != nil {
						code := line[0:ret[0]]
						line = line[ret[1]:]
						lexer.do_in_code(lineno, code)
					}
				}
				break
			}
		}
		if isMatch == false {
			fmt.Errorf("There is some error in proto file with line %v: %v", lineno, line)
			break
		}
	}
	//lexer.start_ss(lineno, line)
	//// check
	//
	//// check: </line>
	//reg := lexer.rules[TAG_LINE_END]
	//if reg.FindStringIndex(line) != nil {
	//	lexer.in_line = false
	//	lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_LINE_END, text: "</line>"})
	//	return
	//}
}

func (lexer *Lexer) do_in_code(lineno int, code string) {
	fmt.Println(lineno, code)
	varStart := lexer.rules[TAG_VAR_START].FindStringIndex(code)
	varEnd := lexer.rules[TAG_VAR_END].FindStringIndex(code)
	if varStart != nil && varEnd != nil {
		varStr := code[varStart[1]:varEnd[0]]
		lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_CODE_START, text: "<code>"})
		lexer.do_in_var(lineno, varStr)
		lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: TAG_CODE_END, text: "</code>"})
	}
}

func (lexer *Lexer) do_in_var(lineno int, varStr string) {
	fmt.Println(lineno, varStr)
}

func (lexer *Lexer) takeToken() *Token {
	if lexer.currTokenIdx >= len(lexer.tokens) {
		return nil
	}
	ret := lexer.tokens[lexer.currTokenIdx]
	lexer.currTokenIdx++
	fmt.Println("takeToken: ", ret.text)
	return ret
}

func (lexer *Lexer) nextTokenType() int {
	if lexer.currTokenIdx >= len(lexer.tokens) {
		return -1
	}
	return lexer.tokens[lexer.currTokenIdx].tokenType
}
