package meta

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"
)

// MockClient é uma implementação fake de Sender + CommentResponder — não faz
// nenhuma chamada de rede. Gera ids no formato real da Meta (wamid.*/mid.*) e
// loga o envio, pra o resto do backend (usecases, reconciliação de status via
// webhook, SSE) poder ser desenvolvido e testado de ponta a ponta antes de
// existir credencial Meta real.
//
// WhatsApp não implementa comments/stories na vida real, então o mock replica
// essa restrição: ReplyPublic/ReplyPrivate falham com ErrUnsupportedOperation
// quando channel == ChannelWhatsApp, mesmo o método existindo no struct.
type MockClient struct {
	channel ChannelType

	mu   sync.Mutex
	sent []SentMessage
}

// SentMessage é o registro de um envio feito pelo mock — usado em testes de
// usecase pra afirmar o que teria sido enviado, sem depender de rede.
type SentMessage struct {
	Kind      string // "text" | "template" | "reply_public" | "reply_private"
	Recipient Recipient
	Body      string
	Template  string
	Params    []string
	SentAt    time.Time
}

var (
	_ Sender           = (*MockClient)(nil)
	_ CommentResponder = (*MockClient)(nil)
)

func NewMockClient(channel ChannelType) *MockClient {
	return &MockClient{channel: channel}
}

func (m *MockClient) Channel() ChannelType { return m.channel }

func (m *MockClient) SendText(ctx context.Context, input SendTextInput) (SendResult, error) {
	return m.record(SentMessage{
		Kind:      "text",
		Recipient: input.Recipient,
		Body:      input.Body,
	})
}

func (m *MockClient) SendTemplate(ctx context.Context, input SendTemplateInput) (SendResult, error) {
	return m.record(SentMessage{
		Kind:      "template",
		Recipient: input.Recipient,
		Template:  input.TemplateName,
		Params:    input.Params,
	})
}

func (m *MockClient) ReplyPublic(ctx context.Context, input ReplyCommentInput) (SendResult, error) {
	if m.channel == ChannelWhatsApp {
		return SendResult{}, ErrUnsupportedOperation
	}
	return m.record(SentMessage{
		Kind:      "reply_public",
		Body:      input.Body,
		Recipient: Recipient{Channel: input.Channel, ExternalID: input.CommentExternalID},
	})
}

func (m *MockClient) ReplyPrivate(ctx context.Context, input ReplyCommentInput) (SendResult, error) {
	if m.channel == ChannelWhatsApp {
		return SendResult{}, ErrUnsupportedOperation
	}
	return m.record(SentMessage{
		Kind:      "reply_private",
		Body:      input.Body,
		Recipient: Recipient{Channel: input.Channel, ExternalID: input.CommentExternalID},
	})
}

// Sent retorna uma cópia do histórico de envios — thread-safe.
func (m *MockClient) Sent() []SentMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]SentMessage, len(m.sent))
	copy(out, m.sent)
	return out
}

func (m *MockClient) record(msg SentMessage) (SendResult, error) {
	msg.SentAt = time.Now()

	m.mu.Lock()
	m.sent = append(m.sent, msg)
	m.mu.Unlock()

	id := fakeMessageID(m.channel)
	if msg.Kind == "template" {
		log.Printf("[meta:mock:%s] template %q -> %s params=%v (id=%s)", m.channel, msg.Template, msg.Recipient.ExternalID, msg.Params, id)
	} else {
		log.Printf("[meta:mock:%s] %s -> %s: %q (id=%s)", m.channel, msg.Kind, msg.Recipient.ExternalID, msg.Body, id)
	}

	return SendResult{MetaMessageID: id, SentAt: msg.SentAt}, nil
}

// fakeMessageID imita o formato real da Meta (wamid.* pro WhatsApp, mid.*
// pro resto) pra não quebrar nenhum parsing/prefixo que o resto do backend já
// vai esperar quando o client real entrar no lugar do mock.
func fakeMessageID(channel ChannelType) string {
	raw := make([]byte, 16)
	_, _ = rand.Read(raw)
	suffix := hex.EncodeToString(raw)

	prefix := "mid"
	if channel == ChannelWhatsApp {
		prefix = "wamid"
	}
	return fmt.Sprintf("%s.MOCK_%s", prefix, suffix)
}
