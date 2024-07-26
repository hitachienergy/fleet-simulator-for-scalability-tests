package hawkbit

import (
	"context"
	"encoding/json"
	"fmt"
	"hitachienergy/scalability-test-client/examples/httppool"
	"io"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/xid"
	"golang.org/x/xerrors"
)

type DMFAmqpService struct {
	tenant string
	id     string

	conn         *amqp.Connection
	ch           *amqp.Channel
	queue        *amqp.Queue
	receiveChann <-chan amqp.Delivery

	exchangeName string
	baseEndpoint string
	ctx          context.Context
	useHttpPool  bool
}

// newDMFAmqpService creates an instance of DMF AMQP Service
func newDMFAmqpService(id string, tenant string, baseEndpoint string, virtualHost string, exchangeName string, useHttpPool bool) (service *DMFAmqpService) {
	service = &DMFAmqpService{
		id:           id,
		tenant:       tenant,
		baseEndpoint: fmt.Sprintf("amqp://%s%s", baseEndpoint, virtualHost),
		exchangeName: exchangeName,
		useHttpPool:  useHttpPool,
	}

	return service
}

// startService starts the update simulation
func (s *DMFAmqpService) startService(ctx context.Context) (err error) {
	s.ctx = ctx
	conn, err := amqp.Dial(s.baseEndpoint)
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	s.conn = conn
	s.ch = ch

	err = s.setupQueue(s.exchangeName, fmt.Sprintf("%s-%s", s.tenant, s.id))
	if err != nil {
		ch.Close()
		s.ch = nil
		conn.Close()
		s.conn = nil
		return err
	}

	// err = ch.Qos(
	// 	1,    // prefetch count
	// 	0,    // prefetch size
	// 	true, // global
	// )
	// if err != nil {
	// 	conn.Close()
	// 	return err
	// }

	return nil
}

// stopService releases AMQP resources
func (s *DMFAmqpService) stopService() (err error) {
	if s.ch != nil {
		s.ch.Close()
		s.ch = nil
	}
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	return nil
}

// reportUpdate reports the update states the the server
func (u *DMFAmqpService) reportUpdate(actionID int64, localStatus LocalUpdateStatus) (err error) {
	feedback := DMFUpdateFeedback{
		ActionID: actionID,
		Messages: localStatus.StatusMsgs,
	}

	switch localStatus.Status {
	case SUCCESSFUL:
		feedback.ActionStatus = UPDATE_FINISHED
	case ERROR:
		feedback.ActionStatus = UPDATE_ERROR
	case DOWNLOADING:
		feedback.ActionStatus = UPDATE_DOWNLOAD
	case DOWNLOADED:
		feedback.ActionStatus = UPDATE_DOWNLOADED
	case RUNNING:
		feedback.ActionStatus = UPDATE_RUNNING
	}

	return u.sendUpdateFeedback(feedback)
}

// createDevice sends the device creation message to the server
func (s *DMFAmqpService) createDevice(info DMFCreate) (err error) {
	properties := amqp.Table{
		AMQP_KEY_TYPE:   AMQP_TYPE_THING_CREATED,
		AMQP_KEY_SENDER: "simulator",
	}

	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return s.sendMessage(data, properties, xid.New().String())
}

// updateAttributes sends attribute update message to the server
func (s *DMFAmqpService) updateAttributes(attributes map[string]string, mode string) error {
	update := DMFAttributesUpdate{
		Attributes: attributes,
		Mode:       mode,
	}
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	properties := amqp.Table{
		AMQP_KEY_TYPE:  AMQP_TYPE_EVENT,
		AMQP_KEY_TOPIC: TOPC_UPDATE_ATTRIBUTES,
	}
	return s.sendMessage(data, properties, xid.New().String())
}

// ping pings the server to check network and show liveness
func (s *DMFAmqpService) ping(correlationID string) (err error) {
	properties := amqp.Table{
		AMQP_KEY_TYPE: AMQP_TYPE_PING,
	}
	return s.sendMessage(nil, properties, correlationID)
}

// sendUpdateFeedback sends a update feedback message to the server
func (s *DMFAmqpService) sendUpdateFeedback(feeback DMFUpdateFeedback) error {
	properties := amqp.Table{
		AMQP_KEY_TYPE:  AMQP_TYPE_EVENT,
		AMQP_KEY_TOPIC: TOPC_UPDATE_ACTION_STATUS,
	}

	data, err := json.Marshal(feeback)
	if err != nil {
		return err
	}
	return s.sendMessage(data, properties, xid.New().String())
}

// setupQueue sets up the exchange and queue that will be used for following communications
func (s *DMFAmqpService) setupQueue(exchangeName string, queueName string) (err error) {
	err = s.ch.ExchangeDeclare(
		exchangeName, // exchange name
		"fanout",     // exchange type
		true,         // durable
		true,         // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	args := make(amqp.Table)
	args["x-message-ttl"] = int64(24 * time.Hour / time.Millisecond)
	args["x-max-length"] = int64(100000)
	q, err := s.ch.QueueDeclare(
		queueName, // name
		false,     // durable
		true,      // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)
	if err != nil {
		return err
	}

	err = s.ch.QueueBind(
		q.Name,       // queue name
		"",           // routing key
		exchangeName, // exchange name
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}
	s.queue = &q

	msgs, err := s.ch.Consume(
		s.queue.Name, // queue name
		"",           // consumer tag
		true,         // auto-acknowledge
		true,         // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}
	s.receiveChann = msgs

	return nil
}

// sendMessage sends a amqp message
func (s *DMFAmqpService) sendMessage(data []byte, properties amqp.Table, correlationtID string) (err error) {
	properties[AMQP_KEY_TENANT] = s.tenant
	properties[AMQP_KEY_THING_ID] = s.id
	// properties[AMQP_KEY_CONTENT_TYPE] = ContentTypeJSON

	message := amqp.Publishing{
		Headers:       properties,
		ContentType:   ContentTypeJSON,
		Body:          data,
		CorrelationId: correlationtID,
		ReplyTo:       s.exchangeName,
	}

	err = s.ch.PublishWithContext(
		s.ctx,
		DMF_EXCHANGE, // exchange name
		"",           // routing key
		false,        // mandatory
		false,        // immediate
		message,
	)
	return err
}

// download downloads firmware from the server using HTTP (weird but this is Hawkbit simulator written as)
func (s *DMFAmqpService) download(ctx context.Context, url string, token string) (body []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Authorization", fmt.Sprintf("TargetToken %s", token))
	res, err := s.sendHTTP(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode > 299 {
		return nil, xerrors.Errorf("Fail to get action with deployment. Status code: %d (%s)", res.StatusCode, res.Status)
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// sendHTTP sends a simple http request. It can either send directly or using HTTP Pool
func (s *DMFAmqpService) sendHTTP(request *http.Request) (*http.Response, error) {
	request.Close = true

	if !s.useHttpPool {
		return httppool.Client.Do(request)
	}
	client := httppool.Pool.Get()
	defer httppool.Pool.Put(client)
	return client.Do(request)
}
