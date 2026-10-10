package outbox

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/go-event-driven/v2/common/log"
	"github.com/ThreeDotsLabs/watermill"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v3/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
)

func NewPostgresPublisher(ctx context.Context, tx *sqlx.Tx) (message.Publisher, error) {
	var publisher message.Publisher

	logger := watermill.NewSlogLogger(log.FromContext(ctx))

	publisher, err := watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create outbox publisher: %w", err)
	}

	// log.CorrelationPublisherDecorator:
	// - extracts correlation ID from msg.Context() and injects it into msg.Metadata
	// - 1st call: for original messages (later: extracted from envelopes and published to Redis)
	// - 2nd call: for the envelopes (published to Postgres's outbox topic)

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}

	publisher = forwarder.NewPublisher(
		publisher,
		forwarder.PublisherConfig{
			ForwarderTopic: outboxTopic,
		},
	)

	publisher = log.CorrelationPublisherDecorator{Publisher: publisher}

	return publisher, nil
}
