package types

type UserRequest struct {
	Type    string `json:"type" validate:"required,oneof=EXTENSION BUG"`
	Email   string `json:"email" validate:"required"`
	Request string `json:"request" validate:"required"`
}
