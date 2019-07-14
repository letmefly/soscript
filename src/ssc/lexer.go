package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

const (
	TOKEN_MIN            = iota
	TOKEN_COMMENT         // //
	TOKEN_EQUAL           // ==
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
	TOKEN_NUMBER
	TOKEN_STRING
	TOKEN_SYMBOL

	TAG_SOSCRIPT_START
	TAG_SOSCRIPT_END

	TAG_DEFAULT_START
	TAG_DEFAULT_END

	TAG_LINE_START
	TAG_LINE_END

	TAG_CODE_START
	TAG_CODE_END

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
	TOKEN_KEYWORD_IF:    `^\s*if\s+`,
	TOKEN_KEYWORD_PRINT: `^\s*print\s+`,
	TOKEN_KEYWORD_AND:   `^\s*&&\s+`,
	TOKEN_KEYWORD_OR:    `^\s*\|\|\s+`,
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
}

func newLexer(fileType string, reader io.Reader) *Lexer {
	lexer := &Lexer{}
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

	for _, v := range lexer.tokens {
		log.Println("line ", v.lineno, v.text)
	}
}

func (lexer *Lexer) parseLine(lineno int, lineText string) {
	line := lineText
	switch lexer.fileType {
	case "ss":
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

	case "not_ss":

	}
}

func (lexer *Lexer) takeToken() *Token {
	if lexer.currTokenIdx >= len(lexer.tokens) {
		return nil
	}
	ret := lexer.tokens[lexer.currTokenIdx]
	lexer.currTokenIdx++
	//fmt.Println("takeToken: ", ret.text)
	return ret
}

func (lexer *Lexer) nextTokenType() int {
	if lexer.currTokenIdx >= len(lexer.tokens) {
		return -1
	}
	return lexer.tokens[lexer.currTokenIdx].tokenType
}
