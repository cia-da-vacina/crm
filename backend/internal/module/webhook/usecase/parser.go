package usecase

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/model"
)

// whatsappPayload segue o shape documentado do webhook da Cloud API
// (object: "whatsapp_business_account").
type whatsappPayload struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Metadata struct {
					PhoneNumberID string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WaID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
				// Statuses é o array de status de entrega/leitura/cobrança
				// (docs oficiais: "status messages webhook reference") — shape
				// inferido da documentação pública, nunca visto de um app real
				// neste ambiente (mesma ressalva de backend/ARCHITECTURE.md §8
				// pros webhooks de engagement). pricing_confirmed só vira true
				// no dia em que isso for validado contra tráfego real.
				Statuses []struct {
					ID        string `json:"id"`
					Status    string `json:"status"`
					Timestamp string `json:"timestamp"`
					Pricing   struct {
						Billable     bool   `json:"billable"`
						PricingModel string `json:"pricing_model"`
						Category     string `json:"category"`
					} `json:"pricing"`
				} `json:"statuses"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

// parseWhatsApp só extrai mensagens de texto — outros tipos (image, audio…)
// são ignorados no MVP texto-first (docs/decisions.md); o campo Type é lido
// só pra decidir esse filtro, o resto do payload de mídia não é parseado.
func parseWhatsApp(raw []byte) ([]model.InboundMessage, error) {
	var payload whatsappPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	var out []model.InboundMessage
	for _, entryItem := range payload.Entry {
		for _, change := range entryItem.Changes {
			names := make(map[string]string, len(change.Value.Contacts))
			for _, c := range change.Value.Contacts {
				names[c.WaID] = c.Profile.Name
			}

			for _, m := range change.Value.Messages {
				if m.Type != "text" {
					continue
				}
				ts, _ := parseUnixSeconds(m.Timestamp)
				out = append(out, model.InboundMessage{
					Channel:       entity.ChannelWhatsApp,
					ExternalID:    m.From,
					DisplayHandle: names[m.From],
					MetaMessageID: m.ID,
					Body:          m.Text.Body,
					Timestamp:     ts,
					PhoneNumberID: change.Value.Metadata.PhoneNumberID,
				})
			}
		}
	}
	return out, nil
}

// parseWhatsAppStatuses extrai o array "statuses" (Frente A do plano de
// adaptação WhatsApp 2026 — núcleo de custo) — eventos sem objeto "pricing"
// preenchido (billable=false e category="") ainda geram um InboundStatus
// pra atualizar o Status da mensagem (sent/delivered/read/failed), só sem
// Category/Billable/PricingModel setados.
func parseWhatsAppStatuses(raw []byte) ([]model.InboundStatus, error) {
	var payload whatsappPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	var out []model.InboundStatus
	for _, entryItem := range payload.Entry {
		for _, change := range entryItem.Changes {
			for _, s := range change.Value.Statuses {
				ts, _ := parseUnixSeconds(s.Timestamp)
				status := model.InboundStatus{
					MetaMessageID: s.ID,
					Status:        s.Status,
					Timestamp:     ts,
				}
				if s.Pricing.Category != "" {
					category := s.Pricing.Category
					billable := s.Pricing.Billable
					pricingModel := s.Pricing.PricingModel
					status.Category = &category
					status.Billable = &billable
					status.PricingModel = &pricingModel
				}
				out = append(out, status)
			}
		}
	}
	return out, nil
}

// messagingPayload é o shape compartilhado por Instagram DM e Facebook
// Messenger — a Meta unificou o formato de webhook de mensageria dos dois
// (object: "instagram" | "page"). reply_to.story e attachments[].type ==
// "story_mention" também chegam por aqui (mesmo array "messaging"), não por
// um campo de webhook separado — é a Meta que unifica assim.
type messagingPayload struct {
	Entry []struct {
		Messaging []struct {
			Sender struct {
				ID string `json:"id"`
			} `json:"sender"`
			Timestamp int64 `json:"timestamp"`
			Message   struct {
				Mid  string `json:"mid"`
				Text string `json:"text"`
				// Echo (mensagem que o próprio agente mandou pela UI nativa
				// da Meta) e mensagens sem "text" (sticker, mídia) são
				// ignoradas — mesmo corte texto-first do WhatsApp.
				IsEcho  bool `json:"is_echo"`
				ReplyTo struct {
					Story struct {
						ID  string `json:"id"`
						URL string `json:"url"`
					} `json:"story"`
				} `json:"reply_to"`
				Attachments []struct {
					Type    string `json:"type"`
					Payload struct {
						URL string `json:"url"`
					} `json:"payload"`
				} `json:"attachments"`
			} `json:"message"`
		} `json:"messaging"`
	} `json:"entry"`
}

// parseMessaging ignora story_reply/story_mention de propósito — esses vão
// pra parseStoryEngagements em vez de virar Message, mesmo payload cru
// (docs/BACKEND-CONTRACT.md §5: story reply/mention é engagement, não
// mensagem 1:1).
func parseMessaging(raw []byte, channel entity.Channel) ([]model.InboundMessage, error) {
	var payload messagingPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	var out []model.InboundMessage
	for _, entryItem := range payload.Entry {
		for _, m := range entryItem.Messaging {
			if m.Message.IsEcho || m.Message.Text == "" {
				continue
			}
			if m.Message.ReplyTo.Story.ID != "" {
				continue
			}
			skip := false
			for _, a := range m.Message.Attachments {
				if a.Type == "story_mention" {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			out = append(out, model.InboundMessage{
				Channel:       channel,
				ExternalID:    m.Sender.ID,
				MetaMessageID: m.Message.Mid,
				Body:          m.Message.Text,
				Timestamp:     time.UnixMilli(m.Timestamp),
			})
		}
	}
	return out, nil
}

// parseStoryEngagements extrai story_reply/story_mention do mesmo payload de
// messaging — shape documentado publicamente pela Meta (Instagram Messaging
// API), não verificado contra um webhook real (sem credencial de app neste
// ambiente — mesma ressalva de pkg/meta, ver backend/ARCHITECTURE.md §6).
func parseStoryEngagements(raw []byte, channel entity.Channel) ([]model.InboundEngagement, error) {
	var payload messagingPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	var out []model.InboundEngagement
	for _, entryItem := range payload.Entry {
		for _, m := range entryItem.Messaging {
			if m.Message.IsEcho {
				continue
			}
			ts := time.UnixMilli(m.Timestamp)

			if m.Message.ReplyTo.Story.ID != "" {
				out = append(out, model.InboundEngagement{
					Channel:          channel,
					Type:             entity.EngagementStoryReply,
					ExternalID:       m.Message.Mid,
					AuthorExternalID: m.Sender.ID,
					Body:             m.Message.Text,
					MediaID:          strPtrOrNil(m.Message.ReplyTo.Story.ID),
					MediaURL:         strPtrOrNil(m.Message.ReplyTo.Story.URL),
					Timestamp:        ts,
				})
				continue
			}

			for _, a := range m.Message.Attachments {
				if a.Type != "story_mention" {
					continue
				}
				out = append(out, model.InboundEngagement{
					Channel:          channel,
					Type:             entity.EngagementStoryMention,
					ExternalID:       m.Message.Mid,
					AuthorExternalID: m.Sender.ID,
					MediaURL:         strPtrOrNil(a.Payload.URL),
					Timestamp:        ts,
				})
			}
		}
	}
	return out, nil
}

// commentsPayload cobre dois shapes distintos de comentário que a Meta
// manda: Instagram usa "changes[].field" == "comments"/"live_comments" com
// value.from/media/id/text; Facebook Page feed usa field == "feed" com
// value.item == "comment" (value.comment_id/post_id/message). Nenhum dos
// dois foi verificado contra um webhook real — mesma ressalva de
// parseStoryEngagements, baseado só na documentação pública da Meta.
type commentsPayload struct {
	Entry []struct {
		Time    int64 `json:"time"`
		Changes []struct {
			Field string `json:"field"`
			Value struct {
				From struct {
					ID string `json:"id"`
				} `json:"from"`
				Media struct {
					ID string `json:"id"`
				} `json:"media"`
				ID   string `json:"id"`
				Text string `json:"text"`
				// Shape do Facebook Page feed (field == "feed").
				Item      string `json:"item"`
				Verb      string `json:"verb"`
				CommentID string `json:"comment_id"`
				PostID    string `json:"post_id"`
				Message   string `json:"message"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

func parseComments(raw []byte, channel entity.Channel) ([]model.InboundEngagement, error) {
	var payload commentsPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	var out []model.InboundEngagement
	for _, entryItem := range payload.Entry {
		ts := time.Now()
		if entryItem.Time > 0 {
			ts = time.Unix(entryItem.Time, 0)
		}

		for _, change := range entryItem.Changes {
			v := change.Value
			switch change.Field {
			case "comments", "live_comments":
				engType := entity.EngagementPostComment
				if change.Field == "live_comments" {
					engType = entity.EngagementLiveComment
				}
				out = append(out, model.InboundEngagement{
					Channel:          channel,
					Type:             engType,
					ExternalID:       v.ID,
					AuthorExternalID: v.From.ID,
					Body:             v.Text,
					MediaID:          strPtrOrNil(v.Media.ID),
					Timestamp:        ts,
				})
			case "feed":
				if v.Item != "comment" || v.Verb != "add" {
					continue
				}
				out = append(out, model.InboundEngagement{
					Channel:          channel,
					Type:             entity.EngagementPostComment,
					ExternalID:       v.CommentID,
					AuthorExternalID: v.From.ID,
					Body:             v.Message,
					MediaID:          strPtrOrNil(v.PostID),
					Timestamp:        ts,
				})
			}
		}
	}
	return out, nil
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseUnixSeconds(raw string) (time.Time, error) {
	seconds, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return time.Now(), err
	}
	return time.Unix(seconds, 0), nil
}
