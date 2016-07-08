package yaml

type lcv struct {
	l, c int
	v    string
}

type sev struct {
	s, e int
	v    string
}

func TransformValues(yaml string, transformer func(string) (string, error)) (string, error) {
	p := newParser([]byte(yaml))
	defer p.destroy()
	parse := p.parse()
	tokenization := p.parser.all_tokens

	lcvs := scanLCVs(parse, nil)
	sevs := scanSEVs(tokenization, lcvs, nil)
	return transformValues(yaml, sevs, transformer)
}

func transformValues(in string, sevs []sev, transformer func(string) (string, error)) (string, error) {
	out := ""

	lastPrinted := 0
	for _, sev := range sevs {
		out += in[lastPrinted:sev.s]
		xformed, err := transformer(sev.v)
		if err != nil {
			return "", err
		}
		out += xformed
		lastPrinted = sev.e
	}

	out += in[lastPrinted:len(in)]
	return out, nil
}

func scanSEVs(tokens []yaml_token_t, lcvs []lcv, sevs []sev) []sev {
	if len(lcvs) == 0 {
		return sevs
	}
	l := lcvs[0].l
	c := lcvs[0].c
	v := lcvs[0].v
	lcvs = lcvs[1:]

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
	sevs = append(sevs, sev{s: token.start_mark.index, e: token.end_mark.index, v: v})

	tokens = tokens[tokenIndex+1:]
	return scanSEVs(tokens, lcvs, sevs)
}

func scanLCVs(n *node, lcvs []lcv) []lcv {
	for idx, ch := range n.children {
		lcvs = scanLCVs(ch, lcvs)
		if nodeIsEncryptable(ch, n, idx) {
			lcvs = append(lcvs, lcv{ch.line, ch.column, ch.value})
		}
	}
	return lcvs
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
