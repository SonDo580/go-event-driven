package message

import (
	"log/slog"
	"tickets/constants"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/lithammer/shortuuid/v3"
)

func useMiddlewares(router *message.Router) {
	router.AddMiddleware(middleware.Recoverer)

	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          router.Logger(),
	}.Middleware)

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			ctx := msg.Context()

			correlationID := msg.Metadata.Get(constants.CorrelationIDMetadataKey)
			if correlationID == "" {
				correlationID = shortuuid.New()
			}

			ctx = log.ContextWithCorrelationID(ctx, correlationID)
			ctx = log.ToContext(ctx, slog.With("correlation_id", correlationID))
			msg.SetContext(ctx)

			return next(msg)
		}
	})

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger := log.FromContext(msg.Context()).With(
				"message_id", msg.UUID,
				"payload", string(msg.Payload),
				"metadata", msg.Metadata,
			)

			logger.Info("Handling a message")

			msgs, err := next(msg)
			if err != nil {
				logger.With("error", err).Error("Error while handling a message")
			}

			return msgs, err
		}
	})
}
