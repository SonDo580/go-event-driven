package message

import (
	"log/slog"
	"tickets/constants"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func useMiddlewares(router *message.Router) {
	router.AddMiddleware(middleware.Recoverer)

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			correlationID := msg.Metadata.Get(constants.MsgMetaCorrelationID)
			if correlationID == "" {
				correlationID = watermill.NewShortUUID()
			}

			ctx := log.ContextWithCorrelationID(msg.Context(), correlationID)
			msg.SetContext(ctx)

			return next(msg)
		}
	})

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger := slog.With(
				"message_id", msg.UUID,
				"payload", string(msg.Payload), // should only do if payload doesn't contain sensitive info
				"metadata", msg.Metadata,
				"handler", message.HandlerNameFromCtx(msg.Context()),
			)

			logger.Info("Handling a message")

			return next(msg)
		}
	})

}
