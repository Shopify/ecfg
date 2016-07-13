package json

import (
	"encoding/json"

	"github.com/Shopify/ecfg/pkg/format"
)

// ExtractPublicKey finds the _public_key value in an ecfg document and
// parses it into a key usable with the crypto library.
func (h *FormatHandler) ExtractPublicKey(data []byte) (key [32]byte, err error) {
	var obj map[string]interface{}
	if err = json.Unmarshal(data, &obj); err != nil {
		return
	}
	return format.ExtractPublicKeyHelper(obj)
}
