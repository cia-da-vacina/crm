package meta

import "fmt"

// Registry resolve qual Sender/CommentResponder usar a partir do ChannelType
// de uma conversa ou engagement. Os usecases dependem só desta interface —
// nunca instanciam um client de canal diretamente, o que torna a troca de
// mock para client real (fase 6 do roadmap) invisível pro resto do backend.
type Registry struct {
	senders map[ChannelType]Sender
}

func NewRegistry() *Registry {
	return &Registry{senders: make(map[ChannelType]Sender)}
}

// Register associa um Sender ao canal que ele declara via Channel(). Chamado
// uma vez por canal na montagem do app (internal/app/app.go).
func (reg *Registry) Register(s Sender) {
	reg.senders[s.Channel()] = s
}

// Sender retorna o client de envio 1:1 configurado para o canal, ou erro se
// nenhum foi registrado (canal desabilitado/sem credenciais em Settings).
func (reg *Registry) Sender(channel ChannelType) (Sender, error) {
	s, ok := reg.senders[channel]
	if !ok {
		return nil, fmt.Errorf("meta: no sender registered for channel %q", channel)
	}
	return s, nil
}

// CommentResponder retorna o mesmo client registrado para o canal, exigindo
// que ele também implemente CommentResponder — só Instagram/Facebook devem.
func (reg *Registry) CommentResponder(channel ChannelType) (CommentResponder, error) {
	s, err := reg.Sender(channel)
	if err != nil {
		return nil, err
	}
	cr, ok := s.(CommentResponder)
	if !ok {
		return nil, fmt.Errorf("meta: channel %q does not support comment engagements: %w", channel, ErrUnsupportedOperation)
	}
	return cr, nil
}
