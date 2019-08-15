package kafka

type Interface interface {
	Consume(topic string, listener func(delivery []byte) error) (err error)
	Publish(topic string, key string, payload []byte) error
	Close()
}
