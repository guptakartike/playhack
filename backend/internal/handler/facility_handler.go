package handler

import (
	"errors"
	"log"
	"net/http"

	"playhack/internal/service"
)

type FacilityHandler struct {
	facilityService *service.FacilityService
}

func NewFacilityHandler(facilityService *service.FacilityService) *FacilityHandler {
	return &FacilityHandler{facilityService: facilityService}
}

func (h *FacilityHandler) HandleListFacilities(w http.ResponseWriter, r *http.Request) {
	facilities, err := h.facilityService.ListFacilities(r.Context())
	if err != nil {
		log.Printf("Error listing facilities: %v", err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, facilities)
}

func (h *FacilityHandler) HandleListCourts(w http.ResponseWriter, r *http.Request) {
	facilityID := r.PathValue("id")

	courts, err := h.facilityService.ListCourtsByFacility(r.Context(), facilityID)
	if err != nil {
		log.Printf("Error listing courts for facility %s: %v", facilityID, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, courts)
}

func (h *FacilityHandler) HandleListSlots(w http.ResponseWriter, r *http.Request) {
	courtID := r.PathValue("id")
	dateStr := r.URL.Query().Get("date")

	slots, err := h.facilityService.ListSlotsByCourtAndDate(r.Context(), courtID, dateStr)
	if err != nil {
		if errors.Is(err, service.ErrInvalidDateFormat) {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		log.Printf("Error listing slots for court %s on date %s: %v", courtID, dateStr, err)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, slots)
}
