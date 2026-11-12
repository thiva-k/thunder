// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package aauth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// sfToken is a Structured Field Token (RFC 9651 Section 3.3.4), kept apart from a String so that
// serialization preserves the distinction.
type sfToken string

// sfParam is one parameter of an item or inner list. Parameters keep their order, which the
// signature base depends on.
type sfParam struct {
	key   string
	value any
}

// sfItem is a bare item with its parameters.
type sfItem struct {
	value  any
	params []sfParam
}

// sfInnerList is an inner list with its parameters.
type sfInnerList struct {
	items  []sfItem
	params []sfParam
}

// sfMember is one dictionary member. Exactly one of item and list is set.
type sfMember struct {
	key  string
	item *sfItem
	list *sfInnerList
}

// sfDictionary is a parsed Structured Field Dictionary, in member order.
type sfDictionary []sfMember

var errSFParse = errors.New("malformed structured field")

// get returns the member with the given key.
func (d sfDictionary) get(key string) (sfMember, bool) {
	for i := len(d) - 1; i >= 0; i-- {
		if d[i].key == key {
			return d[i], true
		}
	}
	return sfMember{}, false
}

// param returns the parameter value with the given key.
func param(params []sfParam, key string) (any, bool) {
	for i := len(params) - 1; i >= 0; i-- {
		if params[i].key == key {
			return params[i].value, true
		}
	}
	return nil, false
}

// parseSFDictionary parses a Structured Field Dictionary (RFC 9651 Section 4.2.2).
func parseSFDictionary(input string) (sfDictionary, error) {
	p := &sfParser{s: input}
	p.skipSP()
	dict := sfDictionary{}
	for !p.eof() {
		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		member := sfMember{key: key}
		if p.peek() == '=' {
			p.i++
			if p.peek() == '(' {
				list, err := p.parseInnerList()
				if err != nil {
					return nil, err
				}
				member.list = list
			} else {
				item, err := p.parseItem()
				if err != nil {
					return nil, err
				}
				member.item = item
			}
		} else {
			params, err := p.parseParams()
			if err != nil {
				return nil, err
			}
			member.item = &sfItem{value: true, params: params}
		}
		dict = append(dict, member)
		p.skipOWS()
		if p.eof() {
			break
		}
		if p.peek() != ',' {
			return nil, errSFParse
		}
		p.i++
		p.skipOWS()
		if p.eof() {
			return nil, errSFParse
		}
	}
	return dict, nil
}

type sfParser struct {
	s string
	i int
}

func (p *sfParser) eof() bool { return p.i >= len(p.s) }

func (p *sfParser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.s[p.i]
}

func (p *sfParser) skipSP() {
	for !p.eof() && p.s[p.i] == ' ' {
		p.i++
	}
}

func (p *sfParser) skipOWS() {
	for !p.eof() && (p.s[p.i] == ' ' || p.s[p.i] == '\t') {
		p.i++
	}
}

func (p *sfParser) parseKey() (string, error) {
	c := p.peek()
	if (c < 'a' || c > 'z') && c != '*' {
		return "", errSFParse
	}
	start := p.i
	for !p.eof() {
		c = p.s[p.i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' || c == '*' {
			p.i++
			continue
		}
		break
	}
	return p.s[start:p.i], nil
}

func (p *sfParser) parseParams() ([]sfParam, error) {
	params := []sfParam{}
	for p.peek() == ';' {
		p.i++
		p.skipSP()
		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		var value any = true
		if p.peek() == '=' {
			p.i++
			value, err = p.parseBareItem()
			if err != nil {
				return nil, err
			}
		}
		params = append(params, sfParam{key: key, value: value})
	}
	return params, nil
}

func (p *sfParser) parseItem() (*sfItem, error) {
	value, err := p.parseBareItem()
	if err != nil {
		return nil, err
	}
	params, err := p.parseParams()
	if err != nil {
		return nil, err
	}
	return &sfItem{value: value, params: params}, nil
}

func (p *sfParser) parseInnerList() (*sfInnerList, error) {
	p.i++ // (
	list := &sfInnerList{items: []sfItem{}}
	for !p.eof() {
		p.skipSP()
		if p.peek() == ')' {
			p.i++
			params, err := p.parseParams()
			if err != nil {
				return nil, err
			}
			list.params = params
			return list, nil
		}
		item, err := p.parseItem()
		if err != nil {
			return nil, err
		}
		list.items = append(list.items, *item)
		if c := p.peek(); c != ' ' && c != ')' {
			return nil, errSFParse
		}
	}
	return nil, errSFParse
}

func (p *sfParser) parseBareItem() (any, error) {
	c := p.peek()
	switch {
	case c == '-' || (c >= '0' && c <= '9'):
		return p.parseNumber()
	case c == '"':
		return p.parseString()
	case c == ':':
		return p.parseByteSequence()
	case c == '?':
		return p.parseBoolean()
	case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '*':
		return p.parseToken(), nil
	default:
		return nil, errSFParse
	}
}

func (p *sfParser) parseNumber() (any, error) {
	start := p.i
	if p.peek() == '-' {
		p.i++
	}
	isDecimal := false
	for !p.eof() {
		c := p.s[p.i]
		if c >= '0' && c <= '9' {
			p.i++
			continue
		}
		if c == '.' && !isDecimal {
			isDecimal = true
			p.i++
			continue
		}
		break
	}
	text := p.s[start:p.i]
	if isDecimal {
		f, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, errSFParse
		}
		return f, nil
	}
	n, err := strconv.ParseInt(text, 10, 64)
	if err != nil || len(strings.TrimPrefix(text, "-")) > 15 {
		return nil, errSFParse
	}
	return n, nil
}

func (p *sfParser) parseString() (string, error) {
	p.i++ // "
	var b strings.Builder
	for !p.eof() {
		c := p.s[p.i]
		p.i++
		switch {
		case c == '\\':
			if p.eof() {
				return "", errSFParse
			}
			next := p.s[p.i]
			if next != '"' && next != '\\' {
				return "", errSFParse
			}
			b.WriteByte(next)
			p.i++
		case c == '"':
			return b.String(), nil
		case c < 0x20 || c > 0x7e:
			return "", errSFParse
		default:
			b.WriteByte(c)
		}
	}
	return "", errSFParse
}

func (p *sfParser) parseToken() sfToken {
	start := p.i
	for !p.eof() {
		c := p.s[p.i]
		if c > 0x20 && c < 0x7f && !strings.ContainsRune(`"(),;<=>?@[\]{}`, rune(c)) {
			p.i++
			continue
		}
		break
	}
	return sfToken(p.s[start:p.i])
}

func (p *sfParser) parseByteSequence() ([]byte, error) {
	p.i++ // :
	end := strings.IndexByte(p.s[p.i:], ':')
	if end < 0 {
		return nil, errSFParse
	}
	encoded := p.s[p.i : p.i+end]
	p.i += end + 1
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// Some signers omit padding; accept that as RFC 9651 Section 4.2.7 recommends.
		decoded, err = base64.RawStdEncoding.DecodeString(strings.TrimRight(encoded, "="))
		if err != nil {
			return nil, errSFParse
		}
	}
	return decoded, nil
}

func (p *sfParser) parseBoolean() (bool, error) {
	p.i++ // ?
	switch p.peek() {
	case '1':
		p.i++
		return true, nil
	case '0':
		p.i++
		return false, nil
	default:
		return false, errSFParse
	}
}

// serializeInnerList serializes an inner list with its parameters (RFC 9651 Section 4.1.1.1).
func serializeInnerList(list *sfInnerList) string {
	parts := make([]string, 0, len(list.items))
	for _, item := range list.items {
		parts = append(parts, serializeBareItem(item.value)+serializeParams(item.params))
	}
	return "(" + strings.Join(parts, " ") + ")" + serializeParams(list.params)
}

func serializeParams(params []sfParam) string {
	var b strings.Builder
	for _, prm := range params {
		b.WriteString(";" + prm.key)
		if v, ok := prm.value.(bool); ok && v {
			continue
		}
		b.WriteString("=" + serializeBareItem(prm.value))
	}
	return b.String()
}

func serializeBareItem(v any) string {
	switch t := v.(type) {
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case string:
		return sfString(t)
	case sfToken:
		return string(t)
	case []byte:
		return ":" + base64.StdEncoding.EncodeToString(t) + ":"
	case bool:
		if t {
			return "?1"
		}
		return "?0"
	default:
		return fmt.Sprint(t)
	}
}

// sfString serializes s as a Structured Field String.
func sfString(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
