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

func TestArchiveFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "url with query",
			raw:  "https://storage.googleapis.com/simba-bq-jdbc-releases/SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip?download=1",
			want: "SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip",
		},
		{
			name: "plain filename",
			raw:  "SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip",
			want: "SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := archiveFilename(tc.raw)
			if got != tc.want {
				t.Fatalf("archiveFilename(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestExcludeOldDrivers_WithQuery(t *testing.T) {
	t.Parallel()

	links := []string{
		"https://storage.googleapis.com/simba-bq-jdbc-releases/SimbaJDBCDriverforGoogleBigQuery42_1.5.4.1008.zip?download=1",
		"https://storage.googleapis.com/simba-bq-jdbc-releases/SimbaJDBCDriverforGoogleBigQuery42_9.9.9.9999.zip?download=1",
	}
	got := excludeOldDrivers(links)

	if len(got) != 1 {
		t.Fatalf("excludeOldDrivers length = %d, want 1", len(got))
	}
	if got[0] != links[1] {
		t.Fatalf("excludeOldDrivers first = %q, want %q", got[0], links[1])
	}
}

func TestIsDownloaded_BackwardCompatible(t *testing.T) {
	tmpDir := t.TempDir()

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir tmp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	oldHistoryLine := "https://storage.googleapis.com/simba-bq-jdbc-releases/SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip"
	if err := os.WriteFile(historyFile, []byte(oldHistoryLine+"\n"), 0o644); err != nil {
		t.Fatalf("write history: %v", err)
	}

	target := "https://example.com/mirror/SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip?x=1"
	got, err := isDownloaded(target)
	if err != nil {
		t.Fatalf("isDownloaded: %v", err)
	}
	if !got {
		t.Fatalf("expected backward-compatible match")
	}
}

func TestAppendToHistory_WritesCanonicalKey(t *testing.T) {
	tmpDir := t.TempDir()

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir tmp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	link := "https://storage.googleapis.com/simba-bq-jdbc-releases/SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip?download=1"
	if err := appendToHistory(link); err != nil {
		t.Fatalf("appendToHistory: %v", err)
	}

	b, err := os.ReadFile(historyFile)
	if err != nil {
		t.Fatalf("ReadFile history: %v", err)
	}
	got := string(b)
	want := "SimbaJDBCDriverforGoogleBigQuery42_1.6.3.1004.zip\n"
	if got != want {
		t.Fatalf("history content = %q, want %q", got, want)
	}
}

func createZipWithFile(zipPath, fileName, body string) (err error) {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	zw := zip.NewWriter(f)
	defer func() {
		if cerr := zw.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	w, err := zw.Create(fileName)
	if err != nil {
		return err
	}

	if _, err = w.Write([]byte(body)); err != nil {
		return err
	}

	return nil
}
