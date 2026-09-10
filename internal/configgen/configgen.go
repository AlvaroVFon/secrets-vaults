// Package configgen generates typed configuration interfaces/structs for a
// consumer from its registered secrets.
package configgen

import (
	"errors"
	"fmt"
	"go/format"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"secrets-vault/internal/secrets"
)

// Language identifies the target language of the generated code.
type Language string

const (
	LangTS Language = "ts"
	LangGo Language = "go"
)

// ErrUnsupportedLanguage is returned when an unknown language is requested.
var ErrUnsupportedLanguage = errors.New("unsupported language")

// ParseLanguage normalizes a raw language string. An empty value defaults to TS.
func ParseLanguage(raw string) (Language, error) {
	switch Language(strings.ToLower(strings.TrimSpace(raw))) {
	case "", LangTS:
		return LangTS, nil
	case LangGo:
		return LangGo, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedLanguage, raw)
	}
}

// Result holds the generated code together with its metadata.
type Result struct {
	Language Language `json:"language"`
	Filename string   `json:"filename"`
	Code     string   `json:"code"`
}

type node struct {
	children map[string]*node
	order    []string
	leaf     *secrets.Secret
}

// Generate builds a typed configuration definition for the given consumer from
// its secrets. Keys using dot notation (e.g. "db.host") are nested. Values are
// never included, only inferred types.
func Generate(consumerName string, items []secrets.Secret, lang Language) (Result, error) {
	if lang != LangTS && lang != LangGo {
		return Result{}, fmt.Errorf("%w: %q", ErrUnsupportedLanguage, lang)
	}

	sorted := make([]secrets.Secret, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Key < sorted[j].Key })

	root := buildTree(sorted)
	name := typeName(consumerName)

	result := Result{Language: lang}
	switch lang {
	case LangGo:
		code, err := generateGo(name, root)
		if err != nil {
			return Result{}, err
		}
		result.Filename = slug(consumerName, "_") + "_config.go"
		result.Code = code
	case LangTS:
		result.Filename = slug(consumerName, "-") + ".config.ts"
		result.Code = generateTS(name, root)
	}

	return result, nil
}

func buildTree(items []secrets.Secret) *node {
	root := &node{children: map[string]*node{}}
	for i := range items {
		secret := items[i]
		cur := root
		for _, part := range splitKey(secret.Key) {
			child, ok := cur.children[part]
			if !ok {
				child = &node{children: map[string]*node{}}
				cur.children[part] = child
				cur.order = append(cur.order, part)
			}
			cur = child
		}
		leaf := secret
		cur.leaf = &leaf
	}
	return root
}

func splitKey(key string) []string {
	raw := strings.Split(key, ".")
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		if part = strings.TrimSpace(part); part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 0 {
		return []string{"field"}
	}
	return parts
}

// inferType guesses a scalar type from a raw string value.
func inferType(value string) string {
	trimmed := strings.TrimSpace(value)
	switch strings.ToLower(trimmed) {
	case "true", "false":
		return "bool"
	}
	if _, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return "int"
	}
	if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return "float64"
	}
	return "string"
}

func goType(inferred string) string {
	switch inferred {
	case "bool", "int", "float64":
		return inferred
	default:
		return "string"
	}
}

func tsType(inferred string) string {
	switch inferred {
	case "bool":
		return "boolean"
	case "int", "float64":
		return "number"
	default:
		return "string"
	}
}

func generateGo(typeName string, root *node) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "type %s struct {\n", typeName)
	writeGoFields(&b, root, 1)
	b.WriteString("}\n")

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return "", fmt.Errorf("format generated go code: %w", err)
	}
	return string(src), nil
}

func writeGoFields(b *strings.Builder, n *node, depth int) {
	indent := strings.Repeat("\t", depth)
	used := map[string]int{}

	for _, key := range n.order {
		child := n.children[key]
		field := uniqueName(goFieldName(key), used)

		if len(child.children) > 0 {
			fmt.Fprintf(b, "%s%s struct {\n", indent, field)
			writeGoFields(b, child, depth+1)
			fmt.Fprintf(b, "%s} `json:\"%s\"`\n", indent, key)
			continue
		}

		typ := "string"
		if child.leaf != nil {
			typ = goType(inferType(child.leaf.Value))
		}
		fmt.Fprintf(b, "%s%s %s `json:\"%s\"`\n", indent, field, typ, key)
	}
}

func generateTS(typeName string, root *node) string {
	var b strings.Builder
	fmt.Fprintf(&b, "export interface %s {\n", typeName)
	writeTSFields(&b, root, 1)
	b.WriteString("}\n")
	return b.String()
}

func writeTSFields(b *strings.Builder, n *node, depth int) {
	indent := strings.Repeat("  ", depth)
	used := map[string]int{}

	for _, key := range n.order {
		child := n.children[key]
		field := uniqueName(tsFieldName(key), used)

		if len(child.children) > 0 {
			fmt.Fprintf(b, "%s%s: {\n", indent, field)
			writeTSFields(b, child, depth+1)
			fmt.Fprintf(b, "%s};\n", indent)
			continue
		}

		typ := "string"
		if child.leaf != nil {
			typ = tsType(inferType(child.leaf.Value))
		}
		fmt.Fprintf(b, "%s%s: %s;\n", indent, field, typ)
	}
}

func typeName(consumerName string) string {
	words := splitWords(consumerName)
	if len(words) == 0 {
		return "ProviderConfig"
	}
	var b strings.Builder
	for _, word := range words {
		b.WriteString(pascalWord(word))
	}
	b.WriteString("Config")
	return b.String()
}

func goFieldName(segment string) string {
	var b strings.Builder
	for _, word := range splitWords(segment) {
		b.WriteString(pascalWord(word))
	}
	return ensureGoIdentifier(b.String())
}

func tsFieldName(segment string) string {
	words := splitWords(segment)
	var b strings.Builder
	for i, word := range words {
		if i == 0 {
			b.WriteString(lowerFirst(word))
		} else {
			b.WriteString(pascalWord(word))
		}
	}
	return ensureTSIdentifier(b.String())
}

func ensureGoIdentifier(name string) string {
	if name == "" {
		return "Field"
	}
	if unicode.IsDigit(rune(name[0])) {
		return "X" + name
	}
	return name
}

func ensureTSIdentifier(name string) string {
	if name == "" {
		return "field"
	}
	if unicode.IsDigit(rune(name[0])) {
		return "_" + name
	}
	return name
}

func splitWords(s string) []string {
	var words []string
	var cur strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(r)
			continue
		}
		if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}
	return words
}

func pascalWord(word string) string {
	if word == "" {
		return word
	}
	if word == strings.ToUpper(word) {
		word = strings.ToLower(word)
	}
	runes := []rune(word)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func lowerFirst(word string) string {
	if word == "" {
		return word
	}
	if word == strings.ToUpper(word) {
		return strings.ToLower(word)
	}
	runes := []rune(word)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func uniqueName(base string, used map[string]int) string {
	if used[base] == 0 {
		used[base] = 1
		return base
	}
	used[base]++
	return fmt.Sprintf("%s%d", base, used[base])
}

func slug(name, sep string) string {
	words := splitWords(name)
	if len(words) == 0 {
		return "provider"
	}
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, sep)
}
