package meta

import (
	"context"
	"errors"
	"testing"
)

func TestMockClient_SendText_RecordsAndReturnsID(t *testing.T) {
	client := NewMockClient(ChannelWhatsApp)

	result, err := client.SendText(context.Background(), SendTextInput{
		Recipient: Recipient{Channel: ChannelWhatsApp, ExternalID: "5551999998888"},
		Body:      "oi!",
	})
	if err != nil {
		t.Fatalf("SendText returned error: %v", err)
	}
	if result.MetaMessageID == "" {
		t.Fatal("expected a non-empty MetaMessageID")
	}

	sent := client.Sent()
	if len(sent) != 1 {
		t.Fatalf("expected 1 recorded message, got %d", len(sent))
	}
	if sent[0].Kind != "text" || sent[0].Body != "oi!" {
		t.Fatalf("unexpected recorded message: %+v", sent[0])
	}
}

func TestMockClient_WhatsApp_RejectsCommentReplies(t *testing.T) {
	client := NewMockClient(ChannelWhatsApp)

	_, err := client.ReplyPublic(context.Background(), ReplyCommentInput{
		Channel:           ChannelWhatsApp,
		CommentExternalID: "irrelevant",
		Body:              "oi",
	})
	if !errors.Is(err, ErrUnsupportedOperation) {
		t.Fatalf("expected ErrUnsupportedOperation, got %v", err)
	}
}

func TestMockClient_Instagram_AllowsCommentReplies(t *testing.T) {
	client := NewMockClient(ChannelInstagram)

	result, err := client.ReplyPrivate(context.Background(), ReplyCommentInput{
		Channel:           ChannelInstagram,
		CommentExternalID: "comment_123",
		Body:              "responde por dm",
	})
	if err != nil {
		t.Fatalf("ReplyPrivate returned error: %v", err)
	}
	if result.MetaMessageID == "" {
		t.Fatal("expected a non-empty MetaMessageID")
	}
}

func TestRegistry_ResolvesSenderPerChannel(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewMockClient(ChannelWhatsApp))
	reg.Register(NewMockClient(ChannelInstagram))

	sender, err := reg.Sender(ChannelWhatsApp)
	if err != nil {
		t.Fatalf("Sender(whatsapp) returned error: %v", err)
	}
	if sender.Channel() != ChannelWhatsApp {
		t.Fatalf("expected whatsapp sender, got channel %q", sender.Channel())
	}

	if _, err := reg.Sender(ChannelFacebook); err == nil {
		t.Fatal("expected error for unregistered channel, got nil")
	}
}

func TestRegistry_CommentResponder_WhatsAppUnsupported(t *testing.T) {
	reg := NewRegistry()
	reg.Register(NewMockClient(ChannelWhatsApp))

	// MockClient implementa o método em struct, então o type-assert em
	// Registry.CommentResponder passa; a rejeição real acontece na chamada,
	// como aconteceria com o client HTTP de verdade (a API da Meta que não
	// existe pro WhatsApp). Cobrimos os dois pontos: aqui a resolução, em
	// TestMockClient_WhatsApp_RejectsCommentReplies a chamada.
	responder, err := reg.CommentResponder(ChannelWhatsApp)
	if err != nil {
		t.Fatalf("CommentResponder(whatsapp) returned error: %v", err)
	}

	_, err = responder.ReplyPublic(context.Background(), ReplyCommentInput{Channel: ChannelWhatsApp})
	if !errors.Is(err, ErrUnsupportedOperation) {
		t.Fatalf("expected ErrUnsupportedOperation, got %v", err)
	}
}
