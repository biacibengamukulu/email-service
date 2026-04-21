package dtos

type SmsSendDto struct {
	Phone   string `json:"phone" validate:"required,phone_number_za"`
	Message string `json:"message" validate:"required,max=160"`
}
