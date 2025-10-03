package main

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	driverDownloadURL = "https://cloud.google.com/bigquery/docs/reference/odbc-jdbc-drivers"
	driverFilename    = "GoogleBigQueryJDBC42.jar"
	downloadsDir      = "downloads"
	historyFile       = "download_history.txt"
	httpTimeout       = 180 * time.Second
)

var excludeDriverList = map[string]struct{}{
	"SimbaJDBCDriverforGoogleBigQuery42_1.5.4.1008.zip":  {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.5.0.1001.zip":  {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.3.3.1004.zip":  {},
	"SimbaBigQueryJDBC42-1.3.2.1003.zip":                 {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.3.0.1001.zip":  {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.25.1029.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.23.1027.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.22.1026.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.21.1025.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.19.1023.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.18.1022.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.16.1020.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.14.1017.zip": {},
	"SimbaJDBCDriverforGoogleBigQuery42_1.2.1.1001.zip":  {},
	"SimbaJDBCDriverforGoogleBigQuery41_1.2.1.1001.zip":  {},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("Fetching driver list page...")
	page, err := fetchPageContent(driverDownloadURL)
	if err != nil {
		return fmt.Errorf("ページ取得に失敗しました: %w", err)
	}

	links := getDriverDownloadLinks(page)
	for i := range links {
		links[i] = normalizeURL(links[i])
	}
	links = excludeOldDrivers(links)

	// Ensure downloads directory exists
	if err := os.MkdirAll(downloadsDir, 0o755); err != nil {
		return fmt.Errorf("出力ディレクトリ作成に失敗: %w", err)
	}

	for _, link := range links {
		if !strings.HasSuffix(strings.ToLower(link), ".zip") {
			continue
		}
		already, err := isDownloaded(link)
		if err != nil {
			return err
		}
		if already {
			fmt.Printf("already downloaded: %s\n", link)
			continue
		}

		fmt.Printf("downloading: %s\n", link)
		zipPath, err := downloadJDBCDriver(link, downloadsDir)
		if err != nil {
			return fmt.Errorf("ダウンロード失敗: %s: %w", link, err)
		}
		fmt.Printf("downloaded: %s\n", zipPath)

		if err := extractSpecificJar(zipPath, downloadsDir); err != nil {
			return fmt.Errorf("JAR 抽出失敗 (%s): %w", link, err)
		}
		if err := appendToHistory(link); err != nil {
			return err
		}
		fmt.Printf("extracted: %s\n", driverFilename)
	}

	return nil
}

func httpClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}

func fetchPageContent(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// getDriverDownloadLinks extracts href values that contain "jdbc".
// For simplicity and zero external deps, use a regex on anchor tags.
func getDriverDownloadLinks(html string) []string {
	// Matches href="..." and captures the value; case-insensitive search for jdbc in the value.
	re := regexp.MustCompile(`(?i)<a[^>]+href\s*=\s*"([^"]+)"[^>]*>`) // "
	matches := re.FindAllStringSubmatch(html, -1)
	var out []string
	for _, m := range matches {
		if len(m) >= 2 {
			href := m[1]
			if strings.Contains(strings.ToLower(href), "jdbc") {
				out = append(out, href)
			}
		}
	}
	return out
}

func normalizeURL(href string) string {
	// 既に絶対 URL ならそのまま
	low := strings.ToLower(href)
	if strings.HasPrefix(low, "http://") || strings.HasPrefix(low, "https://") {
		return href
	}
	// 先頭がスラッシュの相対 URL のみベースドメインを補完
	if strings.HasPrefix(href, "/") {
		return "https://cloud.google.com" + href
	}
	return href
}

func excludeOldDrivers(links []string) []string {
	var out []string
	for _, l := range links {
		base := filepath.Base(l)
		if _, found := excludeDriverList[base]; found {
			continue
		}
		out = append(out, l)
	}
	return out
}

func downloadJDBCDriver(url, destDir string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	filename := filepath.Base(url)
	path := filepath.Join(destDir, filename)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", err
	}

	// Validate zip by trying to open.
	if err := validateZip(path); err != nil {
		return "", err
	}
	return path, nil
}

func validateZip(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	// Ensure close
	_ = r.Close()
	if fi.Size() == 0 {
		return errors.New("downloaded file is empty")
	}
	return nil
}

func extractSpecificJar(zipPath, extractTo string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	var jarFile *zip.File
	for _, f := range zr.File {
		if f.Name == driverFilename {
			jarFile = f
			break
		}
	}
	if jarFile == nil {
		return fmt.Errorf("%s が ZIP 内に見つかりません", driverFilename)
	}

	rc, err := jarFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	base := strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))
	outName := fmt.Sprintf("%s-%s", base, driverFilename)
	outPath := filepath.Join(extractTo, outName)

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		return err
	}
	return nil
}

func isDownloaded(link string) (bool, error) {
	f, err := os.Open(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.TrimSpace(s.Text()) == link {
			return true, nil
		}
	}
	if err := s.Err(); err != nil {
		return false, err
	}
	return false, nil
}

func appendToHistory(link string) error {
	already, err := isDownloaded(link)
	if err != nil {
		return err
	}
	if already {
		return nil
	}
	f, err := os.OpenFile(historyFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(link + "\n"); err != nil {
		return err
	}
	return nil
}
