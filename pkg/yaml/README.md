This package is a one-time export of gopkg.in/yaml.v2, with a few changes:

1. parser (`struct yaml_parser_t`) keeps a full tokenization around instead of
   throwing tokens away as they're consumed.
2. Added `scalar_value_tranformer.go`, which uses the parser and tokenization
   to satisfy `format.FormatHandler`.
