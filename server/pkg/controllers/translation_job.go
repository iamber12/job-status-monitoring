package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"video-translation-status/server/pkg/utils"
)

type JobStatus string

const (
	StatusCompleted JobStatus = "completed"
	StatusError     JobStatus = "error"
	StatusPending   JobStatus = "pending"
)

type Job struct {
	ID         string
	CreatedAt  time.Time
	CompleteAt time.Time
	HasError   bool
}

type TranslationJobHandler struct {
	jobs     sync.Map
	MinDelay time.Duration
	MaxDelay time.Duration
}

func NewTranslationJobHandler(minDelay, maxDelay time.Duration) *TranslationJobHandler {
	return &TranslationJobHandler{
		MinDelay: minDelay,
		MaxDelay: maxDelay,
	}
}

func (t *TranslationJobHandler) CreateJob(c *gin.Context) {
	jobID := uuid.New().String()
	randomDelay := t.MinDelay + time.Duration(rand.Int63n(int64(t.MaxDelay-t.MinDelay)))

	job := &Job{
		ID:         jobID,
		CreatedAt:  time.Now(),
		CompleteAt: time.Now().Add(randomDelay),
		HasError:   time.Now().Second()%2 != 0,
	}

	t.jobs.Store(jobID, job)

	resp := utils.ResponseRenderer("Job created successfully", gin.H{
		"jobID": jobID,
	})
	c.JSON(http.StatusOK, resp)
}

func (t *TranslationJobHandler) GetJobStatus(c *gin.Context) {
	jobData, exists := t.jobs.Load(c.Param("job_id"))
	if !exists {
		resp := utils.ResponseRenderer(fmt.Sprintf("failed to find a job with the given job id: %s", c.Param("id")), nil)
		c.JSON(http.StatusNotFound, resp)
		return
	}

	var status JobStatus
	var httpStatus int
	job := jobData.(*Job)

	if time.Now().Before(job.CompleteAt) {
		status = StatusPending
		httpStatus = http.StatusOK
	} else if job.HasError {
		status = StatusError
		httpStatus = http.StatusInternalServerError
	} else {
		status = StatusCompleted
		httpStatus = http.StatusOK
	}

	resp := utils.ResponseRenderer("Job status", gin.H{
		"status": status,
	})
	c.JSON(httpStatus, resp)
}
