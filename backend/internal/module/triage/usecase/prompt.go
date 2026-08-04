package usecase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/pkg/openai"
)

// triageResponse é o JSON que pedimos pro modelo devolver — ver o prompt em
// buildPrompt. Campos ausentes ficam no zero value do Go (string vazia,
// false, 0), tratados de forma segura pelo caller.
type triageResponse struct {
	Intent            string         `json:"intent"`
	Confidence        float64        `json:"confidence"`
	Summary           string         `json:"summary"`
	InternalNotes     string         `json:"internal_notes"`
	PhoneGateRequired bool           `json:"phone_gate_required"`
	ReadyForHandoff   bool           `json:"ready_for_handoff"`
	Reply             string         `json:"reply"`
	CollectedFields   map[string]any `json:"collected_fields"`
}

func parseTriageResponse(raw string) (triageResponse, error) {
	var r triageResponse
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		return triageResponse{}, fmt.Errorf("invalid JSON from model: %w", err)
	}
	if r.Intent == "" {
		r.Intent = "outro"
	}
	return r, nil
}

// buildPrompt monta system+histórico pro modelo. A regra de phone_gate por
// canal (docs/BACKEND-CONTRACT.md §3: WhatsApp nunca pede telefone, só
// IG/FB quando anônimo e a intenção exige) é explicada pro modelo aqui, mas
// reforçada de forma determinística em RunTriage — o modelo pode errar o
// campo phone_gate_required; a regra de negócio real nunca deixa o backend
// confiar cegamente nisso pra WhatsApp.
func buildPrompt(
	settings entity.AppSettings,
	conv entity.Conversation,
	identification entity.CustomerIdentification,
	messages []entity.Message,
	campaigns []entity.AICampaign,
) []openai.Message {
	var sb strings.Builder

	sb.WriteString("Você é a assistente de triagem da Cia da Vacina, uma rede de clínicas de vacinação. ")
	sb.WriteString("Sua função é cumprimentar o cliente, identificar a intenção dele e coletar informações — ")
	sb.WriteString("você NUNCA substitui um atendente humano numa negociação real, só prepara o contexto pra ele. ")
	sb.WriteString("Intenções possíveis (use exatamente uma destas em \"intent\"): agendar, precos, duvidas, reclamacao, outro. ")
	sb.WriteString("Responda SEMPRE em JSON válido, sem nenhum texto fora do JSON, neste formato exato: ")
	sb.WriteString(`{"intent":"...","confidence":0.0,"summary":"...","internal_notes":"...",` +
		`"phone_gate_required":false,"ready_for_handoff":false,"reply":"...","collected_fields":{}}`)
	sb.WriteString(". \"summary\" é um resumo curto pro atendente que for assumir. ")
	sb.WriteString("\"reply\" é a mensagem que será enviada de volta ao cliente pelo canal dele — português, tom acolhedor e objetivo, sem emoji em excesso.")

	if settings.AISystemPrompt != "" {
		sb.WriteString("\n\nInstruções adicionais configuradas pelo administrador:\n" + settings.AISystemPrompt)
	}
	if settings.AIContext != "" {
		sb.WriteString("\n\nContexto adicional sobre a operação:\n" + settings.AIContext)
	}
	if len(campaigns) > 0 {
		sb.WriteString("\n\nCampanhas ativas no momento (use se forem relevantes pra intenção do cliente):")
		for _, c := range campaigns {
			sb.WriteString(fmt.Sprintf("\n- %s: %s", c.Title, c.Description))
		}
	}

	sb.WriteString(fmt.Sprintf("\n\nCanal desta conversa: %s.", conv.Channel))
	switch {
	case conv.Channel == entity.ChannelWhatsApp:
		sb.WriteString(" Este é o WhatsApp: o número do cliente já foi confirmado pela própria Meta ao entregar a mensagem. NUNCA peça telefone aqui — defina phone_gate_required sempre como false.")
	case identification == entity.IdentificationIdentified:
		sb.WriteString(" O cliente já está identificado (telefone confirmado anteriormente) — não é preciso pedir telefone de novo. Defina phone_gate_required como false.")
	default:
		sb.WriteString(" O cliente ainda é anônimo neste canal (sem telefone confirmado). Se a intenção exigir agendamento ou dados cadastrais, defina phone_gate_required=true e peça o número no seu \"reply\" — a confirmação de posse acontece depois, fora do seu controle. Para dúvidas/preços leves, não peça telefone.")
	}

	msgs := []openai.Message{{Role: "system", Content: sb.String()}}

	for _, m := range messages {
		role := "user"
		if m.Direction == entity.DirectionOut {
			role = "assistant"
		}
		if m.Body == "" {
			continue
		}
		msgs = append(msgs, openai.Message{Role: role, Content: m.Body})
	}

	return msgs
}
