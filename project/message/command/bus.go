package command

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewBus(pub message.Publisher, config cqrs.CommandBusConfig) *cqrs.CommandBus {
	commandBus, err := cqrs.NewCommandBusWithConfig(pub, config)
	if err != nil {
		panic(err)
	}

	return commandBus
}
