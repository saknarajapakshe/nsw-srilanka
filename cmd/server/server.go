package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	engine "github.com/OpenNSW/go-temporal-workflow"
	"github.com/OpenNSW/nsw-task-flow/orchestrator"
)

type server struct {
	manager               *orchestrator.TaskManager
	registry              *orchestrator.TaskTemplateRegistry
	parentWorkflowManager engine.TemporalManager
}

func newServer(manager *orchestrator.TaskManager, registry *orchestrator.TaskTemplateRegistry, parentWorkflowManager engine.TemporalManager) *server {
	return &server{
		manager:               manager,
		registry:              registry,
		parentWorkflowManager: parentWorkflowManager,
	}
}

func (s *server) start(addr string) {
	// Serve the portal UI from static/
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// Serve JSONForms schema files from static/forms/
	http.Handle("/forms/", http.StripPrefix("/forms/", http.FileServer(http.Dir("./static/forms"))))

	// API endpoints
	http.HandleFunc("/api/tasks", s.handleGetTasks)
	http.HandleFunc("/api/start", s.handleStartWorkflow)
	http.HandleFunc("/api/task/", s.handleTaskInteraction)
	http.HandleFunc("/api/template/", s.handleGetTemplate)
	http.HandleFunc("/api/payment/", s.handleGetPaymentDetails)

	log.Printf("[API] Starting HTTP server on %s...", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("[API] HTTP server error: %v", err)
		}
	}()
}

func (s *server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.manager.GetAllTasksRenderInfo()) //nolint:errcheck
}

func (s *server) handleStartWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ApplicantName string `json:"applicant_name"`
		WorkflowType  string `json:"workflow_type"` // "classic" or "npqs"
	}
	json.NewDecoder(r.Body).Decode(&req) //nolint:errcheck

	filePath := "templates/npqs-v2/npqs_workflow.json"
	workflowPrefix := "nsw-phyto-"
	switch req.WorkflowType {
	case "npqs-v1":
		filePath = "templates/npqs/npqs_workflow.json"
		workflowPrefix = "nsw-phyto-v1-"
	case "npqs-v2":
		filePath = "templates/npqs-v2/npqs_workflow.json"
		workflowPrefix = "nsw-phyto-v2-"
	case "fcau":
		filePath = "templates/fcau/fcau_workflow.json"
		workflowPrefix = "nsw-fcau-"
	case "customs":
		filePath = "templates/customs/customs_workflow.json"
		workflowPrefix = "nsw-customs-"
	case "cda":
		filePath = "templates/cda/cda_workflow.json"
		workflowPrefix = "nsw-cda-"
	case "sltb":
		filePath = "templates/sltb/sltb_workflow.json"
		workflowPrefix = "nsw-sltb-"
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read %s", filePath), http.StatusInternalServerError)
		return
	}

	var def engine.WorkflowDefinition
	if err := json.Unmarshal(fileBytes, &def); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse %s", filePath), http.StatusInternalServerError)
		return
	}

	workflowID := workflowPrefix + time.Now().Format("150405")
	log.Printf("[API] Starting Parent workflow %s using %s (applicant=%s)", workflowID, filePath, req.ApplicantName)

	if req.ApplicantName == "" {
		req.ApplicantName = "John Doe"
	}
	initialVars := map[string]any{"applicant_name": req.ApplicantName}

	err = s.parentWorkflowManager.StartWorkflow(context.Background(), workflowID, def, initialVars)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start parent workflow: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "workflow_id": workflowID}) //nolint:errcheck
}

// handleTaskInteraction is the unified endpoint: POST /api/task/{taskID}
// It routes the payload to the correct active Task workflow activity using the stored SubTaskNodeID.
func (s *server) handleTaskInteraction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	taskID := strings.TrimPrefix(r.URL.Path, "/api/task/")
	if taskID == "" {
		http.Error(w, "missing task ID in path", http.StatusBadRequest)
		return
	}

	// The payload from the UI is a namespaced map matching the JSONForms structure,
	// e.g. {"userform": {...}} or {"reviewerform": {...}}
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	err := s.manager.CompleteTaskStep(r.Context(), taskID, payload)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "already completed") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		log.Printf("[API] Failed to complete task step: %v", err)
		http.Error(w, "failed to resume task workflow", http.StatusInternalServerError)
		return
	}

	log.Printf("[API] Task %s resumed successfully", taskID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
}

func (s *server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	templateID := strings.TrimPrefix(r.URL.Path, "/api/template/")
	if templateID == "" {
		http.Error(w, "missing template ID in path", http.StatusBadRequest)
		return
	}

	raw, exists := s.registry.GetGenericTemplate(templateID)
	if !exists {
		http.Error(w, fmt.Sprintf("template %s not found", templateID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(raw) //nolint:errcheck
}

func (s *server) handleGetPaymentDetails(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	taskID := strings.TrimPrefix(r.URL.Path, "/api/payment/")
	if taskID == "" {
		http.Error(w, "missing task ID in path", http.StatusBadRequest)
		return
	}

	record, exists := s.manager.GetTask(taskID)
	if !exists {
		http.Error(w, fmt.Sprintf("task %s not found", taskID), http.StatusNotFound)
		return
	}

	paymentDetails, ok := record.Data["payment_details"]
	if !ok {
		http.Error(w, "no payment details found for task", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(paymentDetails) //nolint:errcheck
}
