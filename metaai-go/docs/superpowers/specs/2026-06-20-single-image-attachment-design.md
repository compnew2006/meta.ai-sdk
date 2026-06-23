# Single Image Attachment Design

## Problem

`AnalyzeImage` selects the captured `attachment_frame.b64` template. That frame
already contains one image attachment. The protobuf walker in
`injectAttachment` skips length-delimited fields incorrectly, fails to detect
the existing `top.f2.f3` attachment, and appends a second attachment. Meta AI
therefore receives both the captured student-card image and the caller's image.

## Considered approaches

1. Correct the protobuf walker and replace the existing attachment in place.
   This preserves the known-good captured frame and changes only the broken
   traversal. This is the selected approach.
2. Use the text-only template and inject a new attachment. This removes the old
   image but changes more bytes in a protocol known to reject small structural
   differences.
3. Capture a new empty attachment template. This depends on another browser
   capture and still leaves the parser bug in place.

## Design

Introduce a shared protobuf field-iteration helper for the message block. It
must advance over varint, fixed32, fixed64, and length-delimited values exactly
once. `injectAttachment` will use it to detect `top.f2.f3`. When the field is
present, the function replaces that field using
`substituteMediaIDInAttachment`; it must never append another attachment.

Malformed protobuf data must return the original frame unchanged rather than
panic or emit a partially corrupted message.

## Test contract

The regression test will load the real `attachment_frame.b64` fixture, build a
frame with a new media ID, and assert all of the following:

- the source message block has exactly one attachment;
- the resulting message block still has exactly one attachment;
- the new media ID, MIME type, and filename are present;
- the captured filename and media ID are absent;
- the text-only template still gains exactly one attachment;
- malformed input does not panic.

The test must fail against the current implementation before production code
changes.

## Verification

After the focused test turns green, run the full Go test suite, race detector,
`go vet`, and `go build`. Then perform one live `AnalyzeImage` call with a plain
test image and confirm that the Meta AI UI shows exactly one attachment.

The separate WebSocket response-timeout behavior is outside this fix.
