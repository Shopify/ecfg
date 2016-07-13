package json

import (
	"github.com/Shopify/ecfg/pkg/format"
)

// FormatHandler simply exposes the methods reqwuired of format.FormatHandler.
type FormatHandler struct{}

var _ format.FormatHandler = &FormatHandler{}
