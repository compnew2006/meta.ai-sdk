package upload

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func imageFile(t *testing.T, name string) string {
	t.Helper()
	p := t.TempDir() + "/" + name
	if err := os.WriteFile(p, []byte("image-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestUploadRejectsInvalidInputs(t *testing.T) {
	img := imageFile(t, "image.jpg")
	cases := []struct {
		name string
		u    Uploader
		path string
		want error
	}{
		{"missing token", Uploader{}, img, ErrMissingToken},
		{"bad token", Uploader{AccessToken: "bad"}, img, ErrBadToken},
		{"missing file", Uploader{AccessToken: "ecto1:x"}, img + "-missing", nil},
		{"non image", Uploader{AccessToken: "ecto1:x"}, imageFile(t, "file.txt"), errors.New("not an image")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.u.Upload(context.Background(), tc.path, 1)
			if err == nil || (tc.want != nil && !strings.Contains(err.Error(), tc.want.Error())) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestUploadReturnsMediaIDAndRequiredHeaders(t *testing.T) {
	u := Uploader{AccessToken: "ecto1:token", CookieHeader: "ecto_1_sess=s", Endpoint: "https://upload.test/path"}
	u.HTTP = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "image-bytes" || r.Header["authorization"][0] != "OAuth ecto1:token" || r.Header["x-entity-name"][0] != "image.jpg" || r.Header["cookie"][0] != "ecto_1_sess=s" {
			t.Fatalf("bad upload request: %#v %q", r.Header, body)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"payload":{"mediaId":"m1"}}`)), Header: make(http.Header)}, nil
	})}
	res, err := u.Upload(context.Background(), imageFile(t, "image.jpg"), 1)
	if err != nil || !res.Success || res.MediaID != "m1" || res.FileSize != 11 || res.MimeType != "image/jpeg" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestUploadSurfacesBoundaryFailures(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		body    string
		network bool
		want    string
	}{
		{"network", 0, "", true, "POST"}, {"precondition", 412, `{"debug_info":{"retriable":false,"type":"Expired","message":"refresh"}}`, false, "Expired"},
		{"server", 500, "boom", false, "500"}, {"structured client", 401, `{"debug_info":{"type":"Denied","message":"no"}}`, false, "Denied"},
		{"plain client", 400, "plain", false, "status 400"}, {"missing media", 200, `{"ok":true}`, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := Uploader{AccessToken: "ecto1:x", Endpoint: "https://upload.test", HTTP: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				if tc.network {
					return nil, errors.New("offline")
				}
				return &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body)), Header: make(http.Header)}, nil
			})}}
			res, err := u.Upload(context.Background(), imageFile(t, "image.jpg"), 1)
			if tc.status == 200 {
				if err != nil || res.Success || res.Error == "" {
					t.Fatalf("res=%+v err=%v", res, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("res=%+v err=%v", res, err)
			}
		})
	}
}
