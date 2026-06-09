package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"orchestrator/internal/modrinth"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// BulkJob tracks the progress of a bulk mod install operation.
type BulkJob struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // pending, running, done, failed
	Completed int       `json:"completed"`
	Total     int       `json:"total"`
	Errors    []string  `json:"errors"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var (
	bulkJobs           = make(map[string]*BulkJob)
	bulkJobsMu         sync.RWMutex
	bulkMaxConcurrency = 3
)

// ModBulkInstallInput is the input for POST /api/mods/bulk.
type ModBulkInstallInput struct {
	Body struct {
		AgentID  string `json:"agentId" doc:"Agent ID"`
		ServerID string `json:"serverId" doc:"Server ID"`
		Items    []struct {
			ProjectID string `json:"projectId"`
			VersionID string `json:"versionId"`
		} `json:"items"`
	}
}

// ModBulkInstallOutput is the output for POST /api/mods/bulk.
type ModBulkInstallOutput struct {
	Body struct {
		JobID string `json:"jobId"`
	}
}

// ModBulkInstall starts a background bulk install of mods.
func (h *Handler) ModBulkInstall(ctx context.Context, input *ModBulkInstallInput) (*ModBulkInstallOutput, error) {
	agent, ok := h.DB.GetAgent(input.Body.AgentID)
	if !ok {
		return nil, huma.Error404NotFound("agent not found", nil)
	}

	jobID := uuid.New().String()
	job := &BulkJob{
		ID:        jobID,
		Status:    "pending",
		Total:     len(input.Body.Items),
		Completed: 0,
		Errors:    []string{},
		UpdatedAt: time.Now().UTC(),
	}

	bulkJobsMu.Lock()
	bulkJobs[jobID] = job
	bulkJobsMu.Unlock()

	go h.runBulkInstall(agent.Host, agent.APIKey, input.Body.ServerID, input.Body.Items, job)

	return &ModBulkInstallOutput{Body: struct {
		JobID string `json:"jobId"`
	}{JobID: jobID}}, nil
}

func (h *Handler) runBulkInstall(agentHost, apiKey, serverID string, items []struct {
	ProjectID string `json:"projectId"`
	VersionID string `json:"versionId"`
}, job *BulkJob) {
	job.Status = "running"
	job.UpdatedAt = time.Now().UTC()

	client := modrinth.NewClient()
	sem := make(chan struct{}, bulkMaxConcurrency)
	var wg sync.WaitGroup

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(projectID, versionID string) {
			defer wg.Done()
			defer func() { <-sem }()

			ver, err := client.GetVersion(versionID)
			if err != nil {
				bulkJobsMu.Lock()
				job.Errors = append(job.Errors, fmt.Sprintf("%s: resolve failed: %v", projectID, err))
				job.UpdatedAt = time.Now().UTC()
				bulkJobsMu.Unlock()
				return
			}

			var fileURL, filename string
			for _, f := range ver.Files {
				if f.Primary {
					fileURL = f.URL
					filename = f.Filename
					break
				}
			}
			if fileURL == "" && len(ver.Files) > 0 {
				fileURL = ver.Files[0].URL
				filename = ver.Files[0].Filename
			}
			if fileURL == "" {
				bulkJobsMu.Lock()
				job.Errors = append(job.Errors, fmt.Sprintf("%s: no download file", projectID))
				job.UpdatedAt = time.Now().UTC()
				bulkJobsMu.Unlock()
				return
			}

			body, err := client.DownloadFile(fileURL)
			if err != nil {
				bulkJobsMu.Lock()
				job.Errors = append(job.Errors, fmt.Sprintf("%s: download failed: %v", projectID, err))
				job.UpdatedAt = time.Now().UTC()
				bulkJobsMu.Unlock()
				return
			}
			defer func() { _ = body.Close() }()

			if err := h.uploadModToAgent(agentHost, apiKey, serverID, filename, body); err != nil {
				bulkJobsMu.Lock()
				job.Errors = append(job.Errors, fmt.Sprintf("%s: upload failed: %v", projectID, err))
				job.UpdatedAt = time.Now().UTC()
				bulkJobsMu.Unlock()
				return
			}

			bulkJobsMu.Lock()
			job.Completed++
			job.UpdatedAt = time.Now().UTC()
			bulkJobsMu.Unlock()
		}(item.ProjectID, item.VersionID)
	}

	wg.Wait()
	bulkJobsMu.Lock()
	if len(job.Errors) > 0 {
		job.Status = "failed"
	} else {
		job.Status = "done"
	}
	job.UpdatedAt = time.Now().UTC()
	bulkJobsMu.Unlock()
	slog.Info("bulk install finished", "job_id", job.ID, "status", job.Status, "completed", job.Completed, "errors", len(job.Errors))
}

// ModBulkJobInput is the input for GET /api/mods/jobs/{job_id}.
type ModBulkJobInput struct {
	JobID string `path:"job_id" doc:"Bulk job ID"`
}

// ModBulkJobOutput is the output for GET /api/mods/jobs/{job_id}.
type ModBulkJobOutput struct {
	Body *BulkJob
}

// ModBulkJob returns the status of a bulk install job.
func (h *Handler) ModBulkJob(ctx context.Context, input *ModBulkJobInput) (*ModBulkJobOutput, error) {
	bulkJobsMu.RLock()
	job, ok := bulkJobs[input.JobID]
	bulkJobsMu.RUnlock()
	if !ok {
		return nil, huma.Error404NotFound("job not found", nil)
	}
	return &ModBulkJobOutput{Body: job}, nil
}
