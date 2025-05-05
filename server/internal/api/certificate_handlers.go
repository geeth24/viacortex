package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"viacortex/internal/db"

	"github.com/go-chi/chi/v5"
)

// getAllCertificates retrieves all certificates
func (h *Handlers) getAllCertificates(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	certificates, err := db.GetAllCertificates(ctx, h.db)
	if err != nil {
		http.Error(w, "Failed to retrieve certificates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(certificates)
}

// getCertificateByID retrieves a certificate by its ID
func (h *Handlers) getCertificateByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	certificate, err := db.GetCertificateByID(ctx, h.db, id)
	if err != nil {
		http.Error(w, "Certificate not found", http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(certificate)
}

// getDomainCertificates retrieves all certificates for a domain
func (h *Handlers) getDomainCertificates(w http.ResponseWriter, r *http.Request) {
	domainIDStr := chi.URLParam(r, "domainID")
	domainID, err := strconv.ParseInt(domainIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid domain ID", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	certificates, err := db.GetCertificatesByDomainID(ctx, h.db, domainID)
	if err != nil {
		http.Error(w, "Failed to retrieve certificates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(certificates)
}

// createCertificate creates a new certificate
func (h *Handlers) createCertificate(w http.ResponseWriter, r *http.Request) {
	var cert db.Certificate
	if err := json.NewDecoder(r.Body).Decode(&cert); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	id, err := db.CreateCertificate(ctx, h.db, &cert)
	if err != nil {
		http.Error(w, "Failed to create certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	cert.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cert)
}

// updateCertificate updates an existing certificate
func (h *Handlers) updateCertificate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
		return
	}
	
	var cert db.Certificate
	if err := json.NewDecoder(r.Body).Decode(&cert); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	cert.ID = id
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	if err := db.UpdateCertificate(ctx, h.db, &cert); err != nil {
		http.Error(w, "Failed to update certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// deleteCertificate deletes a certificate
func (h *Handlers) deleteCertificate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid certificate ID", http.StatusBadRequest)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	if err := db.DeleteCertificate(ctx, h.db, id); err != nil {
		http.Error(w, "Failed to delete certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// getExpiringCertificates retrieves certificates that will expire within a specified number of days
func (h *Handlers) getExpiringCertificates(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30 // Default to 30 days if not specified or invalid
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	certificates, err := db.GetExpiringCertificates(ctx, h.db, days)
	if err != nil {
		http.Error(w, "Failed to retrieve expiring certificates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(certificates)
} 