package crmdemo

import (
	"strings"
	"sync"
	"time"

	"github.com/cia-vacina/crm/backend/internal/domain"
	"github.com/google/uuid"
)

const defaultAISystemPrompt = `Você é a assistente de triagem da Cia da Vacina no WhatsApp.

Como agir:
- Seja cordial, objetiva e em português do Brasil.
- Faça perguntas curtas para entender a necessidade (qual vacina, para quem, unidade preferida).
- Quando houver campanha ativa no contexto, mencione de forma natural se for relevante.
- Ao identificar intenção clara (agendar, preços, dúvidas), resuma e prepare handoff para humano.

O que NÃO fazer:
- Não invente preços fora do contexto/campanhas informados.
- Não confirme agendamento definitivo (isso é do atendente humano).
- Não discuta temas médicos diagnósticos; oriente a falar com o profissional na unidade.
- Não colete dados sensíveis desnecessários (documentos completos, cartão).
- Não prometa prazos que não estão no contexto operacional.`

const defaultAIContext = `Unidades ativas no POC: Centro, Norte, Sul, Leste e Oeste.
Horário de atendimento padrão: seg–sáb 8h–18h.
Foco atual: campanha de vacinação contra gripe e pacotes família.
Após triagem, o atendente humano fecha o agendamento.`

type Store struct {
	mu       sync.RWMutex
	convs    map[uuid.UUID]domain.Conversation
	messages map[uuid.UUID][]domain.Message
	follow   map[uuid.UUID]domain.FollowUp
	pops     []domain.Pop
	reasons  []domain.LossReason
	settings domain.WhatsAppSettings
	units    map[uuid.UUID]domain.Unit
}

func New(units []domain.Unit) *Store {
	s := &Store{
		convs:    map[uuid.UUID]domain.Conversation{},
		messages: map[uuid.UUID][]domain.Message{},
		follow:   map[uuid.UUID]domain.FollowUp{},
		units:    map[uuid.UUID]domain.Unit{},
		settings: domain.WhatsAppSettings{
			WABAID:          "WABA-DEMO-001",
			PhoneNumberID:   "PHONE-CENTRO-001",
			DisplayPhone:    "+55 11 4000-1000",
			TokenMasked:     "EAA•••••••••••••••xyz",
			AIEnabled:       true,
			WebhookVerified: true,
			AISystemPrompt:  defaultAISystemPrompt,
			AIContext:       defaultAIContext,
			AICampaigns: []domain.AICampaign{
				{
					ID:          "camp-gripe-2026",
					Title:       "Campanha Gripe 2026",
					Description: "Vacina da gripe com 15% off em dose individual até o fim do período. Pacote família a partir de R$ 480.",
					StartsOn:    "2026-07-01",
					EndsOn:      "2026-08-31",
					Active:      true,
				},
			},
		},
		reasons: []domain.LossReason{
			{Code: "preco", Label: "Preço elevado"},
			{Code: "concorrente", Label: "Foi para concorrente"},
			{Code: "sem_retorno", Label: "Cliente sem retorno"},
			{Code: "prazo", Label: "Sem disponibilidade de agenda"},
			{Code: "nao_interesse", Label: "Perdeu o interesse"},
			{Code: "outro", Label: "Outro"},
		},
	}
	for _, u := range units {
		s.units[u.ID] = u
	}
	s.seed()
	return s
}

func ptr[T any](v T) *T { return &v }

func (s *Store) seed() {
	now := time.Now().UTC()
	centro := uuid.MustParse("11111111-1111-1111-1111-111111111101")
	norte := uuid.MustParse("11111111-1111-1111-1111-111111111102")
	agent := uuid.MustParse("22222222-2222-2222-2222-222222222202")

	c1 := uuid.MustParse("33333333-3333-3333-3333-333333333301")
	c2 := uuid.MustParse("33333333-3333-3333-3333-333333333302")
	c3 := uuid.MustParse("33333333-3333-3333-3333-333333333303")

	s.convs[c1] = domain.Conversation{
		ID: c1, ContactName: "Maria Silva", ContactPhone: "+5511999990001", UnitID: centro,
		PipelineStage: domain.StageEmAtendimento, Mode: domain.ModeAITriage,
		Intent: ptr("agendar"), AISummary: ptr("Cliente quer agendar vacina da gripe para a filha."),
		LastMessagePreview: "Quero agendar a vacina da gripe", LastMessageAt: now,
	}
	s.convs[c2] = domain.Conversation{
		ID: c2, ContactName: "João Pereira", ContactPhone: "+5511999990002", UnitID: centro,
		PipelineStage: domain.StageEmNegociacao, Mode: domain.ModeHuman, OwnerID: &agent,
		Intent: ptr("precos"), AISummary: ptr("Interessado no pacote família."),
		LastMessagePreview: "Qual o valor do pacote?", LastMessageAt: now.Add(-time.Hour),
	}
	s.convs[c3] = domain.Conversation{
		ID: c3, ContactName: "Ana Costa", ContactPhone: "+5511999990003", UnitID: norte,
		PipelineStage: domain.StageAguardandoFechamento, Mode: domain.ModeHuman, OwnerID: &agent,
		Intent: ptr("agendar"), AISummary: ptr("Combinou valores; falta confirmar horário."),
		LastMessagePreview: "Vou falar com meu marido e retorno", LastMessageAt: now.Add(-24 * time.Hour),
	}

	s.messages[c1] = []domain.Message{
		{ID: uuid.New(), ConversationID: c1, Direction: "in", SenderType: "contact", Body: "Olá, boa tarde!", Status: "read", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: uuid.New(), ConversationID: c1, Direction: "out", SenderType: "ai", Body: "Olá! Sou a assistente da Cia da Vacina. Posso ajudar com agendamento, preços ou dúvidas?", Status: "read", CreatedAt: now.Add(-110 * time.Second)},
		{ID: uuid.New(), ConversationID: c1, Direction: "in", SenderType: "contact", Body: "Quero agendar a vacina da gripe para a minha filha.", Status: "read", CreatedAt: now.Add(-time.Minute)},
		{ID: uuid.New(), ConversationID: c1, Direction: "out", SenderType: "ai", Body: "Perfeito! Vou direcionar você para um atendente da Unidade Centro.", Status: "delivered", CreatedAt: now.Add(-30 * time.Second)},
	}
	s.messages[c2] = []domain.Message{
		{ID: uuid.New(), ConversationID: c2, Direction: "in", SenderType: "contact", Body: "Qual o valor do pacote família?", Status: "read", CreatedAt: now.Add(-time.Hour)},
		{ID: uuid.New(), ConversationID: c2, Direction: "out", SenderType: "agent", Body: "Pacote família a partir de R$ 480. Prefere Unidade Centro?", Status: "read", CreatedAt: now.Add(-55 * time.Minute)},
	}
	s.messages[c3] = []domain.Message{
		{ID: uuid.New(), ConversationID: c3, Direction: "out", SenderType: "agent", Body: "Combinamos amanhã às 10h. Posso confirmar?", Status: "read", CreatedAt: now.Add(-25 * time.Hour)},
		{ID: uuid.New(), ConversationID: c3, Direction: "in", SenderType: "contact", Body: "Vou falar com meu marido e retorno", Status: "read", CreatedAt: now.Add(-24 * time.Hour)},
	}

	f1 := uuid.New()
	s.follow[f1] = domain.FollowUp{
		ID: f1, ConversationID: c3, ContactName: "Ana Costa", ContactPhone: "+5511999990003",
		UnitID: norte, PipelineStage: domain.StageAguardandoFechamento,
		DueAt: now.Add(-time.Hour), Status: "open", Note: "Retomar confirmação de horário",
	}

	s.pops = []domain.Pop{
		{ID: uuid.New(), Title: "Saudação padrão", Body: "Olá! Sou da Cia da Vacina. Em que posso ajudar hoje?", IntentTags: []string{"outro", "duvidas"}, Active: true},
		{ID: uuid.New(), Title: "Agendamento — gripe", Body: "Para agendar a vacina da gripe, preciso do nome completo, data de nascimento e unidade de preferência.", IntentTags: []string{"agendar"}, Active: true},
		{ID: uuid.New(), Title: "Tabela de preços", Body: "Dose individual a partir de R$ 150; pacote família a partir de R$ 480.", IntentTags: []string{"precos"}, Active: true},
		{ID: uuid.New(), Title: "Follow-up fechamento", Body: "Oi! Conseguiu confirmar o horário da vacinação? Posso reservar um encaixe esta semana?", IntentTags: []string{"agendar", "outro"}, Active: true},
	}
}

func (s *Store) ListInbox(unitID *uuid.UUID, stage *domain.PipelineStage) []domain.Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Conversation, 0, len(s.convs))
	for _, c := range s.convs {
		if unitID != nil && c.UnitID != *unitID {
			continue
		}
		if stage != nil && c.PipelineStage != *stage {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (s *Store) GetConversation(id uuid.UUID) (domain.Conversation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.convs[id]
	return c, ok
}

func (s *Store) ListMessages(id uuid.UUID) []domain.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.Message{}, s.messages[id]...)
}

func (s *Store) Claim(id, userID uuid.UUID, userName string) (domain.Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.convs[id]
	if !ok {
		return domain.Conversation{}, ErrNotFound
	}
	if c.OwnerID != nil && *c.OwnerID != userID {
		return domain.Conversation{}, ErrConflict
	}
	c.OwnerID = &userID
	c.Mode = domain.ModeHuman
	s.convs[id] = c
	s.messages[id] = append(s.messages[id], domain.Message{
		ID: uuid.New(), ConversationID: id, Direction: "out", SenderType: "system",
		Body: userName + " assumiu o atendimento.", Status: "sent", CreatedAt: time.Now().UTC(),
	})
	return c, nil
}

func (s *Store) SendMessage(id uuid.UUID, body string) (domain.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.convs[id]
	if !ok {
		return domain.Message{}, ErrNotFound
	}
	if c.Mode == domain.ModeAITriage {
		return domain.Message{}, ErrForbidden
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return domain.Message{}, ErrBadRequest
	}
	msg := domain.Message{
		ID: uuid.New(), ConversationID: id, Direction: "out", SenderType: "agent",
		Body: body, Status: "sent", CreatedAt: time.Now().UTC(),
	}
	s.messages[id] = append(s.messages[id], msg)
	c.LastMessagePreview = body
	c.LastMessageAt = msg.CreatedAt
	s.convs[id] = c
	return msg, nil
}

func (s *Store) UpdatePipeline(id uuid.UUID, stage domain.PipelineStage, reasonCode, reasonText string) (domain.Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.convs[id]
	if !ok {
		return domain.Conversation{}, ErrNotFound
	}
	if stage == domain.StageNaoFechado && reasonCode == "" {
		return domain.Conversation{}, ErrUnprocessable
	}
	c.PipelineStage = stage
	s.convs[id] = c
	if stage == domain.StageAguardandoFechamento || stage == domain.StageNaoFechado {
		note := "Aguardando fechamento — retomar contato"
		if stage == domain.StageNaoFechado {
			note = "Motivo: " + reasonCode
			if reasonText != "" {
				note += " — " + reasonText
			}
		}
		fid := uuid.New()
		s.follow[fid] = domain.FollowUp{
			ID: fid, ConversationID: id, ContactName: c.ContactName, ContactPhone: c.ContactPhone,
			UnitID: c.UnitID, PipelineStage: stage, DueAt: time.Now().UTC().Add(24 * time.Hour),
			Status: "open", Note: note,
		}
	}
	return c, nil
}

func (s *Store) ListFollowUps(unitID *uuid.UUID, status string) []domain.FollowUp {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.FollowUp{}
	for _, f := range s.follow {
		if unitID != nil && f.UnitID != *unitID {
			continue
		}
		if status != "" && f.Status != status {
			continue
		}
		out = append(out, f)
	}
	return out
}

func (s *Store) CompleteFollowUp(id uuid.UUID) (domain.FollowUp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, ok := s.follow[id]
	if !ok {
		return domain.FollowUp{}, ErrNotFound
	}
	f.Status = "done"
	s.follow[id] = f
	return f, nil
}

func (s *Store) ListPops(intent string) []domain.Pop {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []domain.Pop{}
	for _, p := range s.pops {
		if !p.Active {
			continue
		}
		if intent != "" {
			match := false
			for _, t := range p.IntentTags {
				if t == intent {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

func normalizePopTags(tags []string) []string {
	cleaned := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, t := range tags {
		t = strings.TrimSpace(strings.ToLower(t))
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		cleaned = append(cleaned, t)
	}
	if len(cleaned) == 0 {
		return []string{"outro"}
	}
	return cleaned
}

func (s *Store) CreatePop(title, body string, tags []string) (domain.Pop, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" || body == "" {
		return domain.Pop{}, ErrBadRequest
	}
	p := domain.Pop{
		ID:         uuid.New(),
		Title:      title,
		Body:       body,
		IntentTags: normalizePopTags(tags),
		Active:     true,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pops = append([]domain.Pop{p}, s.pops...)
	return p, nil
}

func (s *Store) UpdatePop(id uuid.UUID, title, body string, tags []string) (domain.Pop, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" || body == "" {
		return domain.Pop{}, ErrBadRequest
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.pops {
		if p.ID != id || !p.Active {
			continue
		}
		p.Title = title
		p.Body = body
		p.IntentTags = normalizePopTags(tags)
		s.pops[i] = p
		return p, nil
	}
	return domain.Pop{}, ErrNotFound
}

func (s *Store) DeletePop(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, p := range s.pops {
		if p.ID != id || !p.Active {
			continue
		}
		p.Active = false
		s.pops[i] = p
		return nil
	}
	return ErrNotFound
}

func (s *Store) LossReasons() []domain.LossReason {
	return s.reasons
}

func (s *Store) Settings() domain.WhatsAppSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := s.settings
	if strings.TrimSpace(out.AISystemPrompt) == "" {
		out.AISystemPrompt = defaultAISystemPrompt
	}
	if strings.TrimSpace(out.AIContext) == "" {
		out.AIContext = defaultAIContext
	}
	if out.AICampaigns == nil {
		out.AICampaigns = []domain.AICampaign{}
	}
	return out
}

func (s *Store) UpdateSettings(in domain.WhatsAppSettings, token string) domain.WhatsAppSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	if in.DisplayPhone != "" {
		s.settings.DisplayPhone = in.DisplayPhone
	}
	if in.WABAID != "" {
		s.settings.WABAID = in.WABAID
	}
	if in.PhoneNumberID != "" {
		s.settings.PhoneNumberID = in.PhoneNumberID
	}
	s.settings.AIEnabled = in.AIEnabled
	s.settings.AISystemPrompt = in.AISystemPrompt
	s.settings.AIContext = in.AIContext
	if in.AICampaigns != nil {
		campaigns := make([]domain.AICampaign, 0, len(in.AICampaigns))
		for _, c := range in.AICampaigns {
			c.Title = strings.TrimSpace(c.Title)
			c.Description = strings.TrimSpace(c.Description)
			if c.Title == "" {
				continue
			}
			if c.ID == "" {
				c.ID = uuid.New().String()
			}
			campaigns = append(campaigns, c)
		}
		s.settings.AICampaigns = campaigns
	}
	if token != "" {
		mask := token
		if len(token) > 3 {
			mask = token[:3] + "••••••••"
		}
		s.settings.TokenMasked = mask
	}
	return s.settings
}

func (s *Store) Dashboard(unitID *uuid.UUID) domain.DashboardSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	by := map[domain.PipelineStage]int{
		domain.StageEmAtendimento: 0, domain.StageEmNegociacao: 0,
		domain.StageAguardandoFechamento: 0, domain.StageFechado: 0, domain.StageNaoFechado: 0,
	}
	ai, human, open := 0, 0, 0
	for _, c := range s.convs {
		if unitID != nil && c.UnitID != *unitID {
			continue
		}
		by[c.PipelineStage]++
		if c.Mode == domain.ModeAITriage {
			ai++
		} else {
			human++
		}
		if c.PipelineStage != domain.StageFechado && c.PipelineStage != domain.StageNaoFechado {
			open++
		}
	}
	closed, lost := by[domain.StageFechado], by[domain.StageNaoFechado]
	decided := closed + lost
	rate := 0
	if decided > 0 {
		rate = closed * 100 / decided
	}
	awaiting := 0
	for _, f := range s.follow {
		if f.Status == "open" {
			awaiting++
		}
	}
	units := []domain.DashboardUnitSummary{}
	for id, u := range s.units {
		uo, uc, ul := 0, 0, 0
		for _, c := range s.convs {
			if c.UnitID != id {
				continue
			}
			if c.PipelineStage == domain.StageFechado {
				uc++
			} else if c.PipelineStage == domain.StageNaoFechado {
				ul++
			} else {
				uo++
			}
		}
		d := uc + ul
		ur := 0
		if d > 0 {
			ur = uc * 100 / d
		}
		units = append(units, domain.DashboardUnitSummary{
			UnitID: id, UnitName: u.Name, Open: uo, Closed: uc, ConversionRate: ur,
		})
	}
	return domain.DashboardSummary{
		OpenConversations: open, ByStage: by, Closed: closed, NotClosed: lost,
		ConversionRate: rate, AITriage: ai, Human: human, AwaitingFollowup: awaiting, Units: units,
	}
}
