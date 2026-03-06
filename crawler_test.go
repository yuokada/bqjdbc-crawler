package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		href string
		want string
	}{
		{
			name: "absolute URL is kept",
			href: "https://storage.googleapis.com/simba-bq-jdbc-releases/driver.zip",
			want: "https://storage.googleapis.com/simba-bq-jdbc-releases/driver.zip",
		},
		{
			name: "relative path is resolved",
			href: "/downloads/driver.zip",
			want: "https://cloud.google.com/downloads/driver.zip",
		},
		{
			name: "document relative path is resolved",
			href: "driver.zip",
			want: "https://cloud.google.com/bigquery/docs/reference/driver.zip",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeURL(tc.href)
			if got != tc.want {
				t.Fatalf("normalizeURL(%q) = %q, want %q", tc.href, got, tc.want)
			}
		})
	}
}

func TestIsAllowedDriverDownloadURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "allowed storage host",
			url:  "https://storage.googleapis.com/simba-bq-jdbc-releases/driver.zip",
			want: true,
		},
		{
			name: "allowed cloud host",
			url:  "https://cloud.google.com/downloads/driver.zip",
			want: true,
		},
		{
			name: "allowed subdomain host",
			url:  "https://foo.cloud.google.com/downloads/driver.zip",
			want: true,
		},
		{
			name: "reject http",
			url:  "http://cloud.google.com/downloads/driver.zip",
			want: false,
		},
		{
			name: "reject untrusted host",
			url:  "https://example.com/driver.zip",
			want: false,
		},
		{
			name: "reject malformed",
			url:  "://bad-url",
			want: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isAllowedDriverDownloadURL(tc.url)
			if got != tc.want {
				t.Fatalf("isAllowedDriverDownloadURL(%q) = %v, want %v", tc.url, got, tc.want)
			}
		})
	}
}

func TestExtractSpecificJar_NestedPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "driver.zip")

	if err := createZipWithFile(zipPath, "nested/dir/"+driverFilename, "jar-content"); err != nil {
		t.Fatalf("create zip: %v", err)
	}

	if err := extractSpecificJar(zipPath, tmpDir); err != nil {
		t.Fatalf("extractSpecificJar returned error: %v", err)
	}

	outPath := filepath.Join(tmpDir, "driver-"+driverFilename)
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected extracted jar at %s: %v", outPath, err)
	}
}

func createZipWithFile(zipPath, fileName, body string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	w, err := zw.Create(fileName)
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(body)); err != nil {
		return err
	}

	return nil
}
