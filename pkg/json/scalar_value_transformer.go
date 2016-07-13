// Package json implements functions to load the Public key data from an ecfg
// file, and to walk that data file, encrypting or decrypting any keys which,
// according to the specification, are marked as encryptable (see README.md for
// details).
//
// It may be non-obvious why this is implemented using a scanner and not by
// loading the structure, manipulating it, then dumping it. Since Go's maps are
// explicitly randomized, that would cause the entire structure to be randomized
// each time the file was written, rendering diffs over time essentially
// useless.
package json

import (
	"fmt"

	"github.com/dustin/gojson"
)

// ScalarValueTransformer exposes one method to conform to the
// ecfg.ScalarValueTransformer interface.
type ScalarValueTransformer struct{}

// TransformScalarValues walks a JSON document, replacing all actionable nodes
// with the result of calling the passed-in `action` parameter with the content
// of the node. A node is actionable if it's a string *value* or an array
// element, and its referencing key doesn't begin with an underscore. For each
// actionable node, the contents are replaced with the result of Action.
// Everything else is unchanged, and arbitrary document structure and
// formatting are preserved.
//
// Note that this  underscore-to-disable-encryption syntax does not propagate
// down the hierarchy to children.
// That is:
//   * In {"_a": "b"}, Action will not be run at all.
//   * In {"a": "b"}, Action will be run with "b", and the return value will
//      replace "b".
//   * In {"k": {"a": ["b"]}, Action will run on "b".
//   * In {"_k": {"a": ["b"]}, Action run on "b".
//   * In {"k": {"_a": ["b"]}, Action will not run.
func (svt *ScalarValueTransformer) TransformScalarValues(
	data []byte,
	action func([]byte) ([]byte, error),
) ([]byte, error) {
	var (
		inLiteral    bool
		literalStart int
		isComment    bool
		scanner      json.Scanner
	)
	scanner.Reset()
	pline := newPipeline()
	for i, c := range data {
		switch v := scanner.Step(&scanner, int(c)); v {
		case json.ScanContinue, json.ScanSkipSpace:
			// Uninteresting byte. Just advance to next.
		case json.ScanBeginLiteral:
			inLiteral = true
			literalStart = i
		case json.ScanObjectKey:
			// The literal we just finished reading was a Key. Decide whether it was a
			// encryptable by checking whether the first byte after the '"' was an
			// underscore, then append it verbatim to the output buffer.
			inLiteral = false
			isComment = data[literalStart+1] == '_'
			pline.appendBytes(data[literalStart:i])
		case json.ScanError:
			// Some error happened; just bail.
			pline.flush()
			return nil, fmt.Errorf("invalid json")
		case json.ScanEnd:
			// We successfully hit the end of input.
			return pline.flush()
		default:
			if inLiteral {
				inLiteral = false
				// We finished reading some literal, and it wasn't a Key, meaning it's
				// potentially encryptable. If it was a string, and the most recent Key
				// encountered didn't begin with a '_', we are to encrypt it. In any
				// other case, we append it verbatim to the output buffer.
				if isComment || data[literalStart] != '"' {
					pline.appendBytes(data[literalStart:i])
				} else {
					res := make(chan promiseResult)
					go func(subData []byte) {
						actioned, err := runAction(subData, action)
						res <- promiseResult{actioned, err}
						close(res)
					}(data[literalStart:i])
					pline.appendPromise(res)
				}
			}
		}
		if !inLiteral {
			// If we're in a literal, we save up bytes because we may have to encrypt
			// them. Outside of a literal, we simply append each byte as we read it.
			pline.appendByte(c)
		}
	}
	if scanner.EOF() == json.ScanError {
		// Unexpected EOF => malformed JSON
		pline.flush()
		return nil, fmt.Errorf("invalid json")
	}
	return pline.flush()
}

func runAction(
	data []byte,
	action func([]byte) ([]byte, error),
) ([]byte, error) {
	unquoted, ok := json.UnquoteBytes(data)
	if !ok {
		return nil, fmt.Errorf("invalid json")
	}
	done, err := action(unquoted)
	if err != nil {
		return nil, err
	}
	return quoteBytes(done)
}

// probably a better way to do this, but...
func quoteBytes(in []byte) ([]byte, error) {
	data := []string{string(in)}
	out, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return out[1 : len(out)-1], nil
}
