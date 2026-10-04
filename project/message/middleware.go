package message

import (
	"log/slog"
	"tickets/constants"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func useMiddlewares(router *message.Router, watermillLogger watermill.LoggerAdapter) {
	router.AddMiddleware(middleware.Recoverer)

	// Retry
	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          watermillLogger,
	}.Middleware)

	// Correlation ID
	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			correlationID := msg.Metadata.Get(constants.MsgMetaCorrelationID)
			if correlationID == "" {
				correlationID = watermill.NewShortUUID()
			}

			ctx := log.ToContext(msg.Context(), slog.With("correlation_id", correlationID))
			ctx = log.ContextWithCorrelationID(ctx, correlationID)
			msg.SetContext(ctx)

			return next(msg)
		}
	})

	// Log (must come after the correlation ID middleware)
	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger := log.FromContext(msg.Context()).With(
				"message_id", msg.UUID,
				"payload", string(msg.Payload), // should only do if payload doesn't contain sensitive info
				"metadata", msg.Metadata,
				"handler", message.HandlerNameFromCtx(msg.Context()),
			)

			logger.Info("Handling a message")

			msgs, err := next(msg)
			if err != nil {
				logger.With("error", err).Error("Error while handling a message")
			}

			// must return produced messages (otherwise we will lose them)
			return msgs, err
		}
	})
}
