package toml

import (
	"fmt"
	"strings"

	"github.com/Shopify/ecfg/pkg/format"
)

type FormatHandler struct{}

type encryptableItem struct {
	val   string
	start int
	end   int
}

func (h *FormatHandler) TransformScalarValues(
	toml []byte,
	action func([]byte) ([]byte, error),
) ([]byte, error) {
	var (
		in   = string(toml)
		out  = ""
		prev = 0
	)
	encryptable, err := encryptableItems(in)
	if err != nil {
		return nil, err
	}

	for _, item := range encryptable {
		out += in[prev:item.start]
		val, err := action([]byte(item.val))
		if err != nil {
			return nil, err
		}
		out += fmt.Sprintf("%q", string(val))
		prev = item.end
	}
	out += in[prev:len(in)]

	return []byte(out), nil
}

func encryptableItems(data string) (enc []encryptableItem, err error) {
	lexer := lex(data)

	keyIsNext := false
	suppressEncryption := false

	for {
		item := lexer.nextItem()

		if keyIsNext {
			suppressEncryption = strings.HasPrefix(item.val, "_")
			keyIsNext = false
		}

		switch item.typ {
		case itemKeyStart:
			suppressEncryption = false
			keyIsNext = true
		case itemString, itemRawString, itemMultilineString, itemRawMultilineString:
			if !suppressEncryption {
				enc = append(enc, makeEncryptableItem(item))
			}
		case itemEOF:
			return enc, nil
		case itemError:
			return enc, fmt.Errorf("toml error: %s", item.val)
		}
	}
}

func makeEncryptableItem(lexItem item) encryptableItem {
	p := parser{}
	parsedVal, _ := p.value(lexItem)

	adjustment := 0
	switch lexItem.typ {
	case itemString, itemRawString:
		adjustment = 1
	case itemMultilineString, itemRawMultilineString:
		adjustment = 3
	default:
		panic("bug: invalid item type")
	}

	return encryptableItem{
		val:   parsedVal.(string),
		start: lexItem.start - adjustment,
		end:   lexItem.end + adjustment,
	}
}

// ExtractPublicKey finds the _public_key value in an ecfg document and
// parses it into a key usable with the crypto library.
func (h *FormatHandler) ExtractPublicKey(data []byte) (key [32]byte, err error) {
	var obj map[string]interface{}
	if err = Unmarshal(data, &obj); err != nil {
		return
	}
	return format.ExtractPublicKeyHelper(obj)
}

var _ format.FormatHandler = &FormatHandler{}
