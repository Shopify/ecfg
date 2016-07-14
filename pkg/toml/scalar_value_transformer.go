package toml

import "fmt"

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

	for {
		switch item := lexer.nextItem(); item.typ {
		case itemString, itemRawString, itemMultilineString, itemRawMultilineString:
			enc = append(enc, makeEncryptableItem(item))
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
