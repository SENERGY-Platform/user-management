package kafka

import (
	"errors"
	"github.com/segmentio/kafka-go"
	"github.com/wvanbergen/kazoo-go"
	"io/ioutil"
	"log"
	"sync"
)

type Kafka struct {
	mux        sync.Mutex
	zk         string
	group      string
	broker     []string
	consumers  []*Consumer
	publishers map[string]*Publisher
	debug      bool
}

func Init(zookeeperUrl string, group string, debug bool) (Interface, error) {
	k := Kafka{zk: zookeeperUrl, group: group, debug: debug, publishers: map[string]*Publisher{}}
	var err error
	k.broker, err = GetBroker(zookeeperUrl)
	return &k, err
}

func (this *Kafka) Close() {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, c := range this.consumers {
		c.Stop()
	}
	for _, c := range this.publishers {
		err := c.writer.Close()
		if err != nil {
			log.Println(err)
		}
	}
}

func GetBroker(zk string) (brokers []string, err error) {
	return getBroker(zk)
}

func getBroker(zkUrl string) (brokers []string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	if kz, err := kazoo.NewKazoo(zk, zookeeper); err != nil {
		return brokers, err
	} else {
		return kz.BrokerList()
	}
}

func GetKafkaController(zkUrl string) (controller string, err error) {
	zookeeper := kazoo.NewConfig()
	zookeeper.Logger = log.New(ioutil.Discard, "", 0)
	zk, chroot := kazoo.ParseConnectionString(zkUrl)
	zookeeper.Chroot = chroot
	kz, err := kazoo.NewKazoo(zk, zookeeper)
	if err != nil {
		return controller, err
	}
	controllerId, err := kz.Controller()
	if err != nil {
		return controller, err
	}
	brokers, err := kz.Brokers()
	if err != nil {
		return controller, err
	}
	return brokers[controllerId], err
}

func InitTopic(zkUrl string, topics ...string) (err error) {
	return InitTopicWithConfig(zkUrl, 1, 1, topics...)
}

func InitTopicWithConfig(zkUrl string, numPartitions int, replicationFactor int, topics ...string) (err error) {
	controller, err := GetKafkaController(zkUrl)
	if err != nil {
		log.Println("ERROR: unable to find controller", err)
		return err
	}
	if controller == "" {
		log.Println("ERROR: unable to find controller")
		return errors.New("unable to find controller")
	}
	initConn, err := kafka.Dial("tcp", controller)
	if err != nil {
		log.Println("ERROR: while init topic connection ", err)
		return err
	}
	defer initConn.Close()
	for _, topic := range topics {
		err = initConn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
			ConfigEntries: []kafka.ConfigEntry{
				{ConfigName: "retention.ms", ConfigValue: "-1"},
				{ConfigName: "retention.bytes", ConfigValue: "-1"},
				{ConfigName: "cleanup.policy", ConfigValue: "compact"},
				{ConfigName: "delete.retention.ms", ConfigValue: "86400000"},
				{ConfigName: "segment.ms", ConfigValue: "604800000"},
				{ConfigName: "min.cleanable.dirty.ratio", ConfigValue: "0.1"},
			},
		})
		if err != nil {
			return
		}
	}
	return nil
}
