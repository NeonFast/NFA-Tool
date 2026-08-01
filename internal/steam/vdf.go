package steam

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Map is a simple VDF object (string keys, string or nested map values).
type Map map[string]any

func Dump(m Map) string {
	var b strings.Builder
	writeMap(&b, m, 0)
	return b.String()
}

func writeMap(b *strings.Builder, m Map, indent int) {
	pad := strings.Repeat("\t", indent)
	// Stable-ish order is not required by Steam; range order is fine.
	for k, v := range m {
		switch val := v.(type) {
		case Map:
			b.WriteString(pad)
			b.WriteByte('"')
			b.WriteString(escape(k))
			b.WriteString("\"\n")
			b.WriteString(pad)
			b.WriteString("{\n")
			writeMap(b, val, indent+1)
			b.WriteString(pad)
			b.WriteString("}\n")
		default:
			b.WriteString(pad)
			b.WriteByte('"')
			b.WriteString(escape(k))
			b.WriteString("\"\t\t\"")
			b.WriteString(escape(fmt.Sprint(val)))
			b.WriteString("\"\n")
		}
	}
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// Parse parses a minimal subset of VDF used by Steam config files.
func Parse(data string) (Map, error) {
	p := &parser{s: data}
	p.skipWS()
	return p.parseObject()
}

type parser struct {
	s string
	i int
}

func (p *parser) skipWS() {
	for p.i < len(p.s) {
		c := p.s[p.i]
		if c == '/' && p.i+1 < len(p.s) && p.s[p.i+1] == '/' {
			for p.i < len(p.s) && p.s[p.i] != '\n' {
				p.i++
			}
			continue
		}
		if !unicode.IsSpace(rune(c)) {
			break
		}
		p.i++
	}
}

func (p *parser) parseObject() (Map, error) {
	m := Map{}
	for {
		p.skipWS()
		if p.i >= len(p.s) {
			return m, nil
		}
		if p.s[p.i] == '}' {
			p.i++
			return m, nil
		}
		key, err := p.parseString()
		if err != nil {
			return nil, err
		}
		p.skipWS()
		if p.i >= len(p.s) {
			return nil, fmt.Errorf("unexpected eof after key %q", key)
		}
		if p.s[p.i] == '{' {
			p.i++
			child, err := p.parseObject()
			if err != nil {
				return nil, err
			}
			m[key] = child
			continue
		}
		val, err := p.parseString()
		if err != nil {
			return nil, err
		}
		m[key] = val
	}
}

func (p *parser) parseString() (string, error) {
	p.skipWS()
	if p.i >= len(p.s) {
		return "", fmt.Errorf("unexpected eof")
	}
	if p.s[p.i] != '"' {
		// bare token
		start := p.i
		for p.i < len(p.s) && !unicode.IsSpace(rune(p.s[p.i])) && p.s[p.i] != '{' && p.s[p.i] != '}' {
			p.i++
		}
		return p.s[start:p.i], nil
	}
	p.i++
	var b strings.Builder
	for p.i < len(p.s) {
		c := p.s[p.i]
		if c == '\\' && p.i+1 < len(p.s) {
			p.i++
			b.WriteByte(p.s[p.i])
			p.i++
			continue
		}
		if c == '"' {
			p.i++
			return b.String(), nil
		}
		b.WriteByte(c)
		p.i++
	}
	return "", fmt.Errorf("unterminated string")
}

func asMap(v any) Map {
	if m, ok := v.(Map); ok {
		return m
	}
	return nil
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
}

func mustInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
