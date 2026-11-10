package v1

import (
	"database/sql"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

const cveTestID = "CVE-1999-99999"

type cveTestItem struct {
	transactionID int
	action        string
	pkg           string
	version       string
}

func cveTestRequest(db *sql.DB, cve string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/vulnerabilities/:cve/assets", GetCVEAssets(db))

	req, _ := http.NewRequest("GET", "/v1/vulnerabilities/"+cve+"/assets", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func cveTestCleanup(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM transaction_items WHERE machine_id LIKE 'cve-test-%'",
		"DELETE FROM transactions WHERE machine_id LIKE 'cve-test-%'",
		"DELETE FROM assets WHERE machine_id LIKE 'cve-test-%'",
		"DELETE FROM vulnerabilities WHERE id LIKE 'CVETEST-%'", // cascades to package_vulnerabilities
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Errorf("cleanup %q: %v", stmt, err)
		}
	}
}

// cveTestAsset registers an asset whose items happen one day apart per
// transaction, starting at base.
func cveTestAsset(t *testing.T, db *sql.DB, machine, os string, active bool, base time.Time, items []cveTestItem) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO assets (hostname, machine_id, first_seen, last_seen, is_active, os)
		VALUES ($1, $1, $2, $2, $3, $4)`, machine, base, active, os); err != nil {
		t.Fatal(err)
	}
	for _, it := range items {
		when := base.AddDate(0, 0, it.transactionID)
		if _, err := db.Exec(`INSERT INTO transactions (transaction_id, machine_id, hostname, begin_time, end_time)
			VALUES ($1, $2, $2, $3, $3) ON CONFLICT DO NOTHING`, it.transactionID, machine, when); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO transaction_items (transaction_id, machine_id, action, package, version, release, arch)
			VALUES ($1, $2, $3, $4, $5, '1.el9', 'x86_64')`, it.transactionID, machine, it.action, it.pkg, it.version); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGetCVEAssets_InvalidIdentifier(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	if w := cveTestRequest(db, "not-a-cve"); w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetCVEAssets_UnknownCVE(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()

	if w := cveTestRequest(db, "CVE-1999-00001"); w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestGetCVEAssets_Status(t *testing.T) {
	db := setupReportsTestDB(t)
	defer db.Close()
	cveTestCleanup(t, db)
	defer cveTestCleanup(t, db)

	if _, err := db.Exec(`INSERT INTO vulnerabilities (id, upstream) VALUES ('CVETEST-1', $1)`,
		pq.Array([]string{cveTestID})); err != nil {
		t.Fatal(err)
	}
	for _, p := range [][2]string{{"foo", "1.0"}, {"kernel", "5.0"}, {"kernel", "5.1"}} {
		if _, err := db.Exec(`INSERT INTO package_vulnerabilities (package_name, version, release, vulnerability_id, ecosystem)
			VALUES ($1, $2, '1.el9', 'CVETEST-1', 'AlmaLinux:9')`, p[0], p[1]); err != nil {
			t.Fatal(err)
		}
	}

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	alma := "AlmaLinux 9.4"
	cveTestAsset(t, db, "cve-test-kernel", alma, true, base, []cveTestItem{
		{1, "Install", "kernel", "5.0"},
		{2, "Install", "kernel", "5.1"},
		{3, "Removed", "kernel", "5.0"},
	})
	cveTestAsset(t, db, "cve-test-fixed", alma, true, base, []cveTestItem{
		{1, "Install", "foo", "1.0"},
		{2, "Upgrade", "foo", "2.0"},
		{2, "Upgraded", "foo", "1.0"},
	})
	cveTestAsset(t, db, "cve-test-downgrade", alma, true, base, []cveTestItem{
		{1, "Install", "foo", "1.0"},
		{2, "Upgrade", "foo", "2.0"},
		{2, "Upgraded", "foo", "1.0"},
		{3, "Downgrade", "foo", "1.0"},
		{3, "Downgraded", "foo", "2.0"},
	})
	cveTestAsset(t, db, "cve-test-clean", alma, true, base, []cveTestItem{{1, "Install", "foo", "2.0"}})
	cveTestAsset(t, db, "cve-test-fedora", "Fedora Linux 40", true, base, []cveTestItem{{1, "Install", "foo", "1.0"}})
	cveTestAsset(t, db, "cve-test-inactive", alma, false, base, []cveTestItem{{1, "Install", "foo", "1.0"}})

	w := cveTestRequest(db, "cve-1999-99999") // lowercase input is normalized
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body)
	}
	var resp CVEAssetsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(resp.Advisories, []string{"CVETEST-1"}) {
		t.Errorf("advisories = %v", resp.Advisories)
	}

	got := map[string]CVEAsset{}
	for _, a := range resp.Assets {
		got[a.MachineID] = a
	}

	fixedAt := base.AddDate(0, 0, 2)
	tests := []struct {
		machine  string
		status   string
		packages []string
		since    *time.Time
	}{
		{"cve-test-kernel", "vulnerable", []string{"kernel-5.1-1.el9.x86_64"}, nil},
		{"cve-test-fixed", "fixed", nil, &fixedAt},
		{"cve-test-downgrade", "vulnerable", []string{"foo-1.0-1.el9.x86_64"}, nil},
		{"cve-test-clean", "not_affected", nil, nil},
		{"cve-test-fedora", "unknown", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.machine, func(t *testing.T) {
			a, ok := got[tt.machine]
			if !ok {
				t.Fatal("asset missing from the response")
			}
			if a.Status != tt.status {
				t.Errorf("status = %q, want %q", a.Status, tt.status)
			}
			if !slices.Equal(a.Packages, tt.packages) {
				t.Errorf("packages = %v, want %v", a.Packages, tt.packages)
			}
			switch {
			case (a.Since == nil) != (tt.since == nil):
				t.Errorf("since = %v, want %v", a.Since, tt.since)
			case a.Since != nil && !a.Since.Equal(*tt.since):
				t.Errorf("since = %v, want %v", a.Since, tt.since)
			}
		})
	}

	if _, ok := got["cve-test-inactive"]; ok {
		t.Error("inactive asset must not be listed")
	}
}
