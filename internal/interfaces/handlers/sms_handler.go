package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/biangacila/email-service/internal/application/dtos"
	"github.com/biangacila/email-service/internal/application/services"
	"github.com/biangacila/email-service/pkg/utils"
)

type SmsHandler interface {
	Send(phone, message string)
}

type SmsHandlerImpl struct {
	svc services.SmsService
}

func NewSmsHandler(svc services.SmsService) *SmsHandlerImpl {

	return &SmsHandlerImpl{
		svc: svc,
	}
}

func (h *SmsHandlerImpl) SendPost(w http.ResponseWriter, r *http.Request) {
	// register custom validation
	/*validate := validator.New()
	err := validate.RegisterValidation("phone_number_za", utils.PhoneNumberZA)
	if err != nil {
		panic(err)
	}*/

	payload := dtos.SmsSendDto{}
	// This will reject any fields that are not part of the DTO
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, utils.HttpResponseError(err), http.StatusBadRequest)
		return
	}

	if err := dtos.ValidateAnyWithAnyDto(payload, dtos.SmsSendDto{}); err != nil {
		http.Error(w, utils.HttpResponseError(err), http.StatusBadRequest)
		return
	}

	err := h.svc.Send(payload.Phone, payload.Message)
	if err != nil {
		http.Error(w, utils.HttpResponseError(err), http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"status": "sent successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

func (h *SmsHandlerImpl) SendGet(w http.ResponseWriter, r *http.Request) {
	// Get query params
	phone := r.URL.Query().Get("phone")
	message := r.URL.Query().Get("message")

	// Basic validation
	if phone == "" || message == "" {
		http.Error(w, utils.HttpResponseError(
			fmt.Errorf("phone and message are required"),
		), http.StatusBadRequest)
		return
	}

	// validate phone number
	phone, err := utils.NormalizeSAPhone(phone)
	if err != nil {
		http.Error(w, utils.HttpResponseError(err), http.StatusBadRequest)
		return
	}

	// Optional: reuse DTO validation (recommended for consistency)
	payload := dtos.SmsSendDto{
		Phone:   phone,
		Message: message,
	}

	if err := dtos.ValidateAnyWithAnyDto(payload, dtos.SmsSendDto{}); err != nil {
		http.Error(w, utils.HttpResponseError(err), http.StatusBadRequest)
		return
	}

	// Call service
	err = h.svc.Send(phone, message)
	if err != nil {
		http.Error(w, utils.HttpResponseError(err), http.StatusInternalServerError)
		return
	}

	// Response
	response := map[string]interface{}{
		"status": "sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
