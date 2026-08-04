package model

type CreateUnitRequest struct {
	Name       string  `json:"name"       validate:"required"`
	Code       string  `json:"code"       validate:"required"`
	City       string  `json:"city"       validate:"required"`
	Address    string  `json:"address"    validate:"required"`
	Timezone   string  `json:"timezone"`
	Active     *bool   `json:"active"`
	District   *string `json:"district"`
	Complement *string `json:"complement"`
	Reference  *string `json:"reference"`
}

// UpdateUnitRequest é parcial: campos ausentes (nil) não são alterados.
type UpdateUnitRequest struct {
	Name       *string `json:"name"`
	Code       *string `json:"code"`
	City       *string `json:"city"`
	Address    *string `json:"address"`
	Timezone   *string `json:"timezone"`
	Active     *bool   `json:"active"`
	District   *string `json:"district"`
	Complement *string `json:"complement"`
	Reference  *string `json:"reference"`
}
