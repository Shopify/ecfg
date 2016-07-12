package yaml

type coarseValue struct {
	line, column int
	value        string
}

type preciseValue struct {
	startIndex, endIndex int
	value                string
}

// TransformValues operates in three phases, over the parse tree, then the
// token stream, then the raw text of the yaml.
//
// First, we scan the fully-parsed AST for any nodes which represent scalar
//   values and are either hash values or array elements. These have
//   line/column coordinates, indicating where that value begins.
//   So the output here is a tuple of {line, column, value}.
//
//   f(AST) -> []{line, column, value}
//
// In some cases though, the line/column coordinates don't really get us to
//   exactly the right place, and we don't know where the token ends just from
//   the beginning coordinate of the node. We consult the token stream, seeking
//   to the line/column coordinate of the coarsely-located value from step 1,
//   then, if that doesn't bring us to a SCALAR token, seeking forward until we
//   hit one. This can happen if the document uses YAML Tags (e.g. `a: !!str b`)
//   From this, we generate a new tuple of {start, end, value}, where start and
//   end are the byte offsets into the original document at which the token to
//   replace begins and ends, and value is the parsed, untransformed value.
//
//   g(tokens, []{line, column, value}) -> []{start, end, value}
//
// Finally, we scan through the original document, replacing the segments
//   between {start,end} pairs from step 2 with the result of transforming the
//   associated value according to the `transformer` function passed in here.
//
//   h([]{start, end, value}) -> []{start, end, value}'
//   i(input, []{start, end, value}') -> output
func TransformValues(yaml string, transformer func(string) (string, error)) (string, error) {
	p := newParser([]byte(yaml))
	defer p.destroy()
	parse := p.parse()
	tokenization := p.parser.all_tokens

	var (
		coarseValues  []coarseValue
		preciseValues []preciseValue
	)

	coarseValues = findTransformableValues(parse, nil)
	preciseValues = refineValues(tokenization, coarseValues, nil)

	return transformValues(yaml, preciseValues, transformer)
}

func transformValues(in string, pvalues []preciseValue, transformer func(string) (string, error)) (string, error) {
	out := ""

	lastPrinted := 0
	for _, pvalue := range pvalues {
		out += in[lastPrinted:pvalue.startIndex]
		xformed, err := transformer(pvalue.value)
		if err != nil {
			return "", err
		}
		out += xformed
		lastPrinted = pvalue.endIndex
	}

	out += in[lastPrinted:len(in)]
	return out, nil
}

func refineValues(tokens []yaml_token_t, cvalues []coarseValue, pvalues []preciseValue) []preciseValue {
	if len(cvalues) == 0 {
		return pvalues
	}
	l := cvalues[0].line
	c := cvalues[0].column
	v := cvalues[0].value
	cvalues = cvalues[1:]

	tokenIndex := -1
	matchNextScalar := false
	for index, token := range tokens {
		if token.start_mark.line >= l && token.start_mark.column >= c {
			matchNextScalar = true
		}
		if matchNextScalar && token.typ == yaml_SCALAR_TOKEN {
			tokenIndex = index
			break
		}
	}

	token := tokens[tokenIndex]
	pvalues = append(pvalues, preciseValue{startIndex: token.start_mark.index, endIndex: token.end_mark.index, value: v})

	tokens = tokens[tokenIndex+1:]
	return refineValues(tokens, cvalues, pvalues)
}

func findTransformableValues(n *node, cvalues []coarseValue) []coarseValue {
	for idx, ch := range n.children {
		cvalues = findTransformableValues(ch, cvalues)
		if nodeIsEncryptable(ch, n, idx) {
			cvalues = append(cvalues, coarseValue{ch.line, ch.column, ch.value})
		}
	}
	return cvalues
}

func nodeIsEncryptable(n, parent *node, index int) bool {
	switch parent.kind {
	case sequenceNode:
		return n.kind == scalarNode
	case mappingNode:
		return n.kind == scalarNode && index%2 == 1
	default:
		return false
	}
}