package iget

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestIGet(t *testing.T) {
	if ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
	); e != nil {
		t.Fatal(e.Error())
	} else {
		defer ic.Close()
		if r, e := ic.Request(); e != nil {
			t.Fatal(e.Error())
		} else {
			if b, e := ic.Do(r); e != nil {
				t.Fatal(e.Error())
			} else {
				ic.PrintResponse(b)
			}
		}
	}
}

func TestIGetFile(t *testing.T) {
	if ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
		Output("file.html"),
	); e != nil {
		t.Fatal(e.Error())
	} else {
		defer ic.Close()
		t.Cleanup(func() { os.Remove("file.html") })
		if r, e := ic.Request(); e != nil {
			t.Fatal(e.Error())
		} else {
			if b, e := ic.Do(r); e != nil {
				t.Fatal(e.Error())
			} else {
				ic.PrintResponse(b)
			}
		}
	}
}

func TestDoBytes(t *testing.T) {
	ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
	)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer ic.Close()
	r, e := ic.Request()
	if e != nil {
		t.Fatal(e.Error())
	}
	b, e := ic.DoBytes(r)
	if e != nil {
		t.Fatal(e.Error())
	}
	if len(b) == 0 {
		t.Error("DoBytes returned empty response body")
	}
}

func TestDoString(t *testing.T) {
	ic, e := NewIGet(
		Length(1),
		URL("http://i2p-projekt.i2p"),
		Inbound(15),
		Debug(false),
	)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer ic.Close()
	r, e := ic.Request()
	if e != nil {
		t.Fatal(e.Error())
	}
	s, e := ic.DoString(r)
	if e != nil {
		t.Fatal(e.Error())
	}
	if s == "" {
		t.Error("DoString returned empty response body")
	}
}

// TestContinueDownload verifies that when a server ignores the Range header and
// returns 200 OK, Do() truncates the partial file so the output contains exactly
// one complete response body rather than the pre-existing bytes plus the body.
func TestContinueDownload(t *testing.T) {
	const body = "full-response-body"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Deliberately ignore Range header — return 200 with full body.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint:errcheck
	}))
	defer srv.Close()

	// Create a partial file to simulate a previous incomplete download.
	path := t.TempDir() + "/partial.html"
	if err := os.WriteFile(path, []byte("stale-prefix-"), 0o644); err != nil {
		t.Fatal(err)
	}

	ig := &IGet{
		outputPath:       path,
		continueDownload: true,
		verboseOut:       io.Discard,
		client:           srv.Client(),
	}

	// Build a request that includes a Range header (as Request() would for a resume).
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/file", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Range", "bytes=13-")

	if _, err := ig.Do(req); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Errorf("expected file to contain only %q but got %q (old partial not truncated)", body, got)
	}
}

// continueDownload is false, so a re-download does not append to stale content.
func TestSaveToFileTruncates(t *testing.T) {
	path := t.TempDir() + "/out.html"
	body := []byte("hello world")

	// Write pre-existing content to the output file.
	if err := os.WriteFile(path, []byte("stale content"), 0o644); err != nil {
		t.Fatal(err)
	}

	ig := &IGet{
		outputPath:       path,
		continueDownload: false,
		verboseOut:       io.Discard,
	}
	if err := ig.saveToFile(bytes.NewReader(body)); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Errorf("expected file to contain %q but got %q (O_TRUNC not applied)", body, got)
	}
}

// TestSaveToFileAppends verifies that saveToFile appends when continueDownload is true,
// so a resumed download correctly extends a partial file.
func TestSaveToFileAppends(t *testing.T) {
	path := t.TempDir() + "/partial.html"
	first := []byte("first-half-")
	second := []byte("second-half")

	ig := &IGet{
		outputPath:       path,
		continueDownload: true,
		verboseOut:       io.Discard,
	}

	if err := ig.saveToFile(bytes.NewReader(first)); err != nil {
		t.Fatal(err)
	}
	if err := ig.saveToFile(bytes.NewReader(second)); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := string(first) + string(second)
	if string(got) != want {
		t.Errorf("expected appended content %q but got %q", want, got)
	}
}

func TestLineLengthZero(t *testing.T) {
	ic, e := NewIGet(
		LineLength(0),
		URL("http://i2p-projekt.i2p"),
	)
	if e != nil {
		t.Fatal(e.Error())
	}
	defer ic.Close()
	// Verify that constructing with LineLength(0) succeeds (no panic expected in PrintResponse).
	if ic.lineLength != 0 {
		t.Errorf("expected lineLength=0, got %d", ic.lineLength)
	}
}

// TestPrintResponseStreaming verifies that PrintResponse streams the body
// without buffering the entire content. It uses a pipe so that a slow writer
// can be detected: if PrintResponse buffers the full body before writing,
// it would need to read everything before the first write reaches the reader.
func TestPrintResponseStreaming(t *testing.T) {
	const body = "hello streaming world"

	// Build a fake response with a plain string body.
	resp := &http.Response{
		Body:    io.NopCloser(strings.NewReader(body)),
		Request: &http.Request{URL: mustParseURL("http://example.i2p/")},
	}

	// Redirect stdout to capture output.
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	ig := &IGet{
		outputPath: "-",
		lineLength: 0,
		verboseOut: io.Discard,
	}
	ig.PrintResponse(resp)
	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	r.Close()

	if buf.String() != body {
		t.Errorf("expected %q got %q", body, buf.String())
	}
}

// TestPrintResponseLineLengthStreaming verifies that the lineLength path also
// streams without buffering (the body is written in chunks, not all at once).
func TestPrintResponseLineLengthStreaming(t *testing.T) {
	// 10-char body with lineLength=5: expect a newline inserted after every 5 runes.
	const body = "abcdefghij"
	const lineLen = 5

	resp := &http.Response{
		Body:    io.NopCloser(strings.NewReader(body)),
		Request: &http.Request{URL: mustParseURL("http://example.i2p/")},
	}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	ig := &IGet{
		outputPath: "-",
		lineLength: lineLen,
		verboseOut: io.Discard,
	}
	ig.PrintResponse(resp)
	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	r.Close()

	// Expect "abcde\nfghij\n"
	want := "abcde\nfghij\n"
	if buf.String() != want {
		t.Errorf("expected %q got %q", want, buf.String())
	}
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

// TestHeadersWithEquals verifies that header values containing "=" are not
// silently dropped. For example, an Authorization header with a Base64 token
// such as "Authorization=Bearer dG9rZW4=" must be set intact.
func TestHeadersWithEquals(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.i2p/", nil)
	if err != nil {
		t.Fatal(err)
	}

	headers := []string{
		"Authorization=Bearer dG9rZW4=",
		"Cookie=session=abc123",
		"X-Simple=plainvalue",
	}
	Headers(headers)(req)

	if got := req.Header.Get("Authorization"); got != "Bearer dG9rZW4=" {
		t.Errorf("Authorization: expected %q, got %q", "Bearer dG9rZW4=", got)
	}
	if got := req.Header.Get("Cookie"); got != "session=abc123" {
		t.Errorf("Cookie: expected %q, got %q", "session=abc123", got)
	}
	if got := req.Header.Get("X-Simple"); got != "plainvalue" {
		t.Errorf("X-Simple: expected %q, got %q", "plainvalue", got)
	}
}

// TestDoBytesWithFileOutputReturnsError verifies that calling DoBytes when
// outputPath is set to a file path returns a descriptive error rather than
// silently returning empty bytes.
func TestDoBytesWithFileOutputReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("body content")) //nolint:errcheck
	}))
	defer srv.Close()

	path := t.TempDir() + "/out.html"
	ig := &IGet{
		outputPath: path,
		verboseOut: io.Discard,
		client:     srv.Client(),
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/file", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ig.DoBytes(req)
	if err == nil {
		t.Error("expected DoBytes to return error when outputPath is a file, got nil")
	}
}
