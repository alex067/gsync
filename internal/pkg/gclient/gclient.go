package gclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alex067/gsync/internal/pkg/gcontext"
)

var ErrCleanShutdown = fmt.Errorf("shutdown signal")
var ErrNotThatSerious = fmt.Errorf("internal request failure but try again")
var ErrInternalFailure = fmt.Errorf("internal request failure")

// What we expect from Grafana API
type GrafanaDashboard struct {
	Meta      map[string]interface{} `json:"meta"`
	Dashboard map[string]interface{} `json:"dashboard"`
}

type GrafanaDashboardClient struct {
	Mutex              sync.Mutex
	FilePath           string
	Dashboard          GrafanaDashboard
	FolderUid          string
	LastVersion        int
	IsDashboardChanged bool
	Uid                string
}

type GrafanaClient struct {
	Url        string
	TenantId   string
	ApiKey     string
	Interval   time.Duration
	HttpClient *http.Client
	Logger     *slog.Logger
}

// Sets default request headers to authenticate to Grafana
func (gc *GrafanaClient) setRequestHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gc.ApiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grafana-Org-Id", gc.TenantId)
}

func (gc *GrafanaClient) createRequest(apiUrl string, method string, payload []byte) (*http.Response, error) {
	var req *http.Request
	var err error
	if payload != nil {
		req, err = http.NewRequest(method, apiUrl, bytes.NewBuffer(payload))
	} else {
		req, err = http.NewRequest(method, apiUrl, nil)
	}
	if err != nil {
		gc.Logger.Error(
			"error creating request",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	gc.setRequestHeaders(req)

	resp, err := gc.HttpClient.Do(req)
	if err != nil {
		gc.Logger.Error(
			"error making request",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return resp, nil
}

// Generates random string for temp dashboard resource
func (gc *GrafanaClient) generateRandomUid() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randUid := make([]byte, 14)
	for i := range randUid {
		randUid[i] = letters[rand.Intn(len(letters))]
	}
	return string(randUid)
}

// Creates temp dashboard to watch over for changes
// Dashboards are prefixed with hash and recorded in local disk
func (gc *GrafanaClient) generateTempDashboard(dashboardFilePath, folderUid string) (string, error) {
	apiUrl := fmt.Sprintf("%s/api/dashboards/db", gc.Url)
	var dashboard map[string]interface{}

	// Ignore error since file is validated
	dashboardFileData, _ := os.ReadFile(dashboardFilePath)
	if err := json.Unmarshal(dashboardFileData, &dashboard); err != nil {
		return "", err
	}

	// Generate random string hash
	newUid := gc.generateRandomUid()

	// Overwrite uid and append preview title
	dashboardTitle := dashboard["title"]
	dashboard["uid"] = newUid
	dashboard["title"] = fmt.Sprintf("%s (Gsync %s)", dashboardTitle, newUid)
	dashboard["version"] = 0
	dashboard["id"] = nil
	dashboard["description"] = fmt.Sprintf("Generated by gsync. Watcher for %s", dashboardTitle)

	requestBody := make(map[string]interface{})
	requestBody["dashboard"] = dashboard
	requestBody["overwrite"] = false

	if folderUid != "" {
		requestBody["folderUid"] = folderUid
		requestBody["message"] = fmt.Sprintf("Gsync preview dashboard for %s", dashboardTitle)
	}

	payload, _ := json.Marshal(requestBody)

	resp, err := gc.createRequest(apiUrl, "POST", payload)
	if err != nil {
		gc.Logger.Error(
			"error creating request",
			slog.String("error", err.Error()),
		)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status=%d, body=%s", resp.StatusCode, string(body))
	}

	return newUid, nil
}

// Simply checks if the dashboard exists in Grafana
func (gc *GrafanaClient) isDashboardExist(uid string) (bool, error) {
	apiUrl := fmt.Sprintf("%s/api/dashboards/uid/%s", gc.Url, uid)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		gc.Logger.Error(
			"error creating request",
			slog.String("error", err.Error()))
		return false, ErrInternalFailure
	}

	gc.setRequestHeaders(req)

	resp, err := gc.HttpClient.Do(req)
	if err != nil {
		gc.Logger.Error(
			"error making request",
			slog.String("error", err.Error()))
		return false, ErrInternalFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

func (gcd *GrafanaDashboardClient) setAndCompareDashboardVersion() {
	gcd.Mutex.Lock()
	defer gcd.Mutex.Unlock()
	if gcd.LastVersion != 0 {
		gcd.IsDashboardChanged = gcd.LastVersion != int(gcd.Dashboard.Meta["version"].(float64))
	}
	gcd.LastVersion = int(gcd.Dashboard.Meta["version"].(float64))
}

// Main watcher process
// Deploys temp Grafana resource and watches for version changes
// Version changes trigger a process to save changes to disk
func (gc *GrafanaClient) StartWatchingDashboard(
	ctx context.Context,
	configContext gcontext.GConfigContext,
	dbClient *GrafanaDashboardClient,
) error {
	// Check for existing watcher dashboards
	watcherUid := configContext.GetResourceByPath(dbClient.FilePath)
	if watcherUid == "" {
		gc.Logger.Info("Creating watcher dashboard...")
		// Deploy temp dashboard to watch
		watcherUid, err := gc.generateTempDashboard(dbClient.FilePath, dbClient.FolderUid)
		if err != nil {
			gc.Logger.Error(
				"error creating dashboard",
				slog.String("error", err.Error()))
			return err
		}
		// Record new dashboard UID in local config file
		configContext.SetNewResource(watcherUid, dbClient.FilePath)
		dbClient.Uid = watcherUid
		gc.Logger.Info(
			"Watcher dashboard created",
			slog.String("url", fmt.Sprintf("%s/d/%s", gc.Url, watcherUid)))
	} else {
		// Check if dashboard manually deleted by user
		isDashboardExist, err := gc.isDashboardExist(watcherUid)
		if err != nil {
			gc.Logger.Error(
				"error checking for existing dashboard",
				slog.String("uid", watcherUid),
				slog.String("path", dbClient.FilePath),
				slog.String("error", err.Error()))
			return err
		}

		if !isDashboardExist {
			gc.Logger.Info("Error fetching watcher dashboard from Grafana")
			gc.Logger.Info("Creating watcher dashboard...")

			watcherUid, err = gc.generateTempDashboard(dbClient.FilePath, dbClient.FolderUid)
			if err != nil {
				gc.Logger.Error(
					"error creating dashboard",
					slog.String("error", err.Error()))
				return err
			}
			configContext.SetNewResource(watcherUid, dbClient.FilePath)
			dbClient.Uid = watcherUid
			gc.Logger.Info(
				"Watcher dashboard created",
				slog.String("url", fmt.Sprintf("%s/d/%s", gc.Url, watcherUid)))
		} else {
			dbClient.Uid = watcherUid
			gc.Logger.Info(
				"Watcher dashboard found",
				slog.String("url", fmt.Sprintf("%s/d/%s", gc.Url, watcherUid)))
		}
	}

	// Listen to interrupt and term signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	maxRetry := 3
	retry := 0

	// Start polling timer
	ticker := time.NewTicker(gc.Interval)
	defer ticker.Stop()

	gc.Logger.Info("Watching...")

	for {
		select {
		case <-ticker.C:
			// make request
			if err := gc.GetDashboardChanges(dbClient); err != nil {
				if err != ErrNotThatSerious {
					return err
				}
				gc.Logger.Info("error detected, attempting retry...", slog.Int("retry", retry))
				if retry >= maxRetry {
					gc.Logger.Error("max retries reached")
					return ErrInternalFailure
				} else {
					retry += 1
				}
			}
			if dbClient.IsDashboardChanged {
				gc.Logger.Info("Version change detected, saving changes...")
				if err := gc.SaveChangesToDisk(dbClient); err != nil {
					gc.Logger.Error(err.Error())
				}
			}
		case sig := <-signals:
			gc.Logger.Info(fmt.Sprintf("Received signal: %v", sig))
			return ErrCleanShutdown
		case <-ctx.Done():
			gc.Logger.Info("Context cancelled")
			return ctx.Err()
		}
	}
}

// Fetches dashboard schema at intervals to watch for any changes
// Detected changes are saved in memory
func (gc *GrafanaClient) GetDashboardChanges(dbClient *GrafanaDashboardClient) error {
	apiUrl := fmt.Sprintf("%s/api/dashboards/uid/%s", gc.Url, dbClient.Uid)

	resp, err := gc.createRequest(apiUrl, "GET", nil)
	if err != nil {
		gc.Logger.Error(
			"error creating request",
			slog.String("error", err.Error()),
		)
		return ErrInternalFailure
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		gc.Logger.Error(
			"error fetching dashboard version",
			slog.Int("status", resp.StatusCode),
			slog.String("error", string(body)),
		)
		return ErrNotThatSerious
	}

	if err := json.Unmarshal(body, &dbClient.Dashboard); err != nil {
		gc.Logger.Error(
			"error reading response body",
			slog.String("error", err.Error()),
		)
		return ErrInternalFailure
	}

	dbClient.setAndCompareDashboardVersion()
	return nil
}

// Saves current state of dashboard to local json file
func (gc *GrafanaClient) SaveChangesToDisk(dbClient *GrafanaDashboardClient) error {
	var dashboard map[string]interface{}

	dashboardFileData, _ := os.ReadFile(dbClient.FilePath)
	if err := json.Unmarshal(dashboardFileData, &dashboard); err != nil {
		return err
	}

	versionIncrement := dashboard["version"]
	// Check for changed status again to avoid unnecessary version increments
	if dbClient.IsDashboardChanged {
		versionIncrement = dashboard["version"].(float64) + 1
	}

	dbClient.Dashboard.Dashboard["id"] = dashboard["id"]
	dbClient.Dashboard.Dashboard["uid"] = dashboard["uid"]
	dbClient.Dashboard.Dashboard["title"] = dashboard["title"]
	dbClient.Dashboard.Dashboard["version"] = versionIncrement
	dbClient.Dashboard.Dashboard["description"] = dashboard["description"]

	dashboardJson, _ := json.MarshalIndent(dbClient.Dashboard.Dashboard, "", "\t")
	err := os.WriteFile(dbClient.FilePath, dashboardJson, 0644)
	if err != nil {
		return err
	}
	dbClient.IsDashboardChanged = false
	return nil
}

func (gc *GrafanaClient) DeleteWatcherDashboard(dbClient *GrafanaDashboardClient) error {
	apiUrl := fmt.Sprintf("%s/api/dashboards/uid/%s", gc.Url, dbClient.Uid)
	resp, err := gc.createRequest(apiUrl, "DELETE", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s", string(body))
	}

	return nil
}