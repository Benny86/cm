// +build !windows

package rewriter

import (
	"bytes"
	"fmt"
	"testing"
)

var clearSequence = fmt.Sprintf("%c[%dA%c[2K\r", 27, 1, 27)

// TestWriterPosix by writing and flushing many times. The output buffer
// must contain the clearCursor and clearLine sequences.
func TestWriterPosix(t *testing.T) {
	out := new(bytes.Buffer)
	w := New(out)

	for _, tcase := range []struct {
		input, expectedOutput string
	}{
		{input: "foo\n", expectedOutput: "foo\n"},
		{input: "bar\n", expectedOutput: "foo\n" + clearSequence + "bar\n"},
		{input: "fizz\n", expectedOutput: "foo\n" + clearSequence + "bar\n" + clearSequence + "fizz\n"},
	} {
		t.Run(tcase.input, func(t *testing.T) {
			w.Write([]byte(tcase.input))
			w.Flush()
			output := out.String()
			if output != tcase.expectedOutput {
				t.Fatalf("want %q, got %q", tcase.expectedOutput, output)
			}
		})
	}
}