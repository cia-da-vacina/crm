package model

import (
	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	usermodel "github.com/cia-da-vacina/crm/backend/internal/module/user/model"
)

// Me = User (embutido, achatado no JSON) + units — MeResponse do openapi.yaml
// (allOf User + {units}). unit_ids (bare) e units (objetos completos)
// coexistem de propósito: o primeiro já vem do embed de usermodel.User.
type Me struct {
	usermodel.User
	Units []entity.Unit `json:"units"`
}
