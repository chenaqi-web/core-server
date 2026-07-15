package kafka

import (
	"context"

	"github.com/IBM/sarama"
)

type MessageHandler func(ctx context.Context, msg *sarama.ConsumerMessage) error

type SimpleConsumer struct {
	ready   chan bool
	handler MessageHandler
}

func (consumer *SimpleConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *SimpleConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *SimpleConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			if err := consumer.handler(session.Context(), msg); err == nil {
				session.MarkMessage(msg, "")
				session.Commit()
			}

		case <-session.Context().Done():
			return session.Context().Err()
		}
	}
}
