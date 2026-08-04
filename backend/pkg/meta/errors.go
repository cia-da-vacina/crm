package meta

import "errors"

// ErrUnsupportedOperation é retornado quando o canal não implementa a
// operação pedida — ex.: WhatsApp não tem posts/comments/stories, só a Cloud
// API de mensagens 1:1, então ReplyPublic/ReplyPrivate falham nesse canal.
var ErrUnsupportedOperation = errors.New("meta: operation not supported for this channel")
