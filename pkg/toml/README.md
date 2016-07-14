This package is a one-time export of https://github.com/BurntSushi/toml, with a few changes:

1. Tokens (lexer output, `struct item`) are enhanced with start and end indexes
   into the input file.
2. Added `scalar_value_tranformer.go`, which uses the lexer API to satisfy
   `format.FormatHandler`.
