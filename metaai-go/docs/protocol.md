# Clippy wire protocol

The chat transport uses Meta AI's binary Clippy WebSocket protocol. This module's
wire-format contract comes from browser frames captured on 2026-06-19.

The implementation keeps the captured envelope structure and substitutes only
request-specific values such as message text, conversation ID, request ID, and
an optional media attachment. Golden fixtures under
`internal/clippy/testdata/` lock the encoder to that captured structure.

Key invariants:

- SEND frames use subtype `0x00`.
- Protobuf fields stay inside the captured envelope nesting.
- The frame header preserves the observed `body + 2` length convention.
- Image analysis injects the media attachment into the captured attachment field.
- Connection headers and query parameters match the browser transport.

Run `go test ./internal/clippy` after changing framing, parsing, or templates.
