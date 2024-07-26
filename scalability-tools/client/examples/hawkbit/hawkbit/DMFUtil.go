package hawkbit

import "sync"

var DMF_EXCHANGE = "dmf.exchange"

const (
	TOPC_UPDATE_ACTION_STATUS       = "UPDATE_ACTION_STATUS"
	TOPIC_DOWNLOAD_AND_INSTALL      = "DOWNLOAD_AND_INSTALL"
	TOPIC_CANCEL_DOWNLOAD           = "CANCEL_DOWNLOAD"
	TOPC_UPDATE_ATTRIBUTES          = "UPDATE_ATTRIBUTES"
	TOPC_UPDATE_AUTO_CONFIRM        = "UPDATE_AUTO_CONFIRM"
	TOPIC_DOWNLOAD                  = "DOWNLOAD"
	TOPIC_REQUEST_ATTRIBUTES_UPDATE = "REQUEST_ATTRIBUTES_UPDATE"
	TOPIC_MULTI_ACTION              = "MULTI_ACTION"
	TOPC_BATCH_DOWNLOAD             = "BATCH_DOWNLOAD"
	TOPC_BATCH_DOWNLOAD_AND_INSTALL = "BATCH_DOWNLOAD_AND_INSTALL"
	TOPIC_CONFIRM                   = "CONFIRM"
)

const (
	AMQP_TYPE_EVENT         = "EVENT"
	AMQP_TYPE_THING_CREATED = "THING_CREATED"
	AMQP_TYPE_THING_DELETED = "THING_DELETED"
	AMQP_TYPE_THING_REMOVED = "THING_REMOVED"
	AMQP_TYPE_PING          = "PING"
	AMQP_TYPE_PING_RESPONSE = "PING_RESPONSE"
)

const AMQP_KEY_TYPE = "type"
const AMQP_KEY_TENANT = "tenant"
const AMQP_KEY_THING_ID = "thingId"
const AMQP_KEY_SENDER = "sender"
const AMQP_KEY_TOPIC = "topic"
const AMQP_KEY_CONTENT_TYPE = "content-type"

const ContentTypeJSON = "application/json"

const (
	UPDATE_DOWNLOAD        = "DOWNLOAD"
	UPDATE_RETRIEVED       = "RETRIEVED"
	UPDATE_RUNNING         = "RUNNING"
	UPDATE_FINISHED        = "FINISHED"
	UPDATE_ERROR           = "ERROR"
	UPDATE_WARNING         = "WARNING"
	UPDATE_CANCELED        = "CANCELED"
	UPDATE_CANCEL_REJECTED = "CANCEL_REJECTED"
	UPDATE_DOWNLOADED      = "DOWNLOADED"
	UPDATE_CONFIRMED       = "CONFIRMED"
	UPDATE_DENIED          = "DENIED"
)

const (
	ATTRIBUTES_MODE_MERGE   = "MERGE"
	ATTRIBUTES_MODE_REPLACE = "REPLACE"
	ATTRIBUTES_MODE_REMOVE  = "REMOVE"
)

type DMFUpdateFeedback struct {
	ActionID         int64    `json:"actionId"`
	SoftwareModuleId int64    `json:"softwareModuleId"`
	ActionStatus     string   `json:"actionStatus"`
	Messages         []string `json:"message"`
}

type DMFMultiAction struct {
	Topic  string    `json:"topic"`
	Weight int64     `json:"weight"`
	Action DMFAction `json:"action"`
}

type DMFAction struct {
	ID                  int64            `json:"actionId"`
	TargetSecurityToken string           `json:"targetSecurityToken"`
	SoftwareModules     []SoftwareModule `json:"softwareModules"`
}

type SoftwareModule struct {
	ID        int64         `json:"moduleId"`
	Type      string        `json:"moduleType"`
	Version   string        `json:"moduleVersion"`
	Artifacts []DMFArtifact `json:"artifacts"`
}

type DMFArtifact struct {
	Filename string            `json:"filename"`
	Urls     map[string]string `json:"urls"`
	Hashes   map[string]string `json:"hashes"`
	Size     int64             `json:"size"`
}

type DMFCancel struct {
	ActionID int64 `json:"actionId"`
}

type DMFAttributesUpdate struct {
	Attributes map[string]string `json:"attributes"`
	Mode       string            `json:"mode"`
}

type DMFCreate struct {
	Name            string              `json:"name"`
	AttributeUpdate DMFAttributesUpdate `json:"attributeUpdate"`
}

// --------------------------------------------------

// pingTracer records all ping request that is sent out and yet to be replied
type pingTracer struct {
	pingID map[string]struct{}
	*sync.Mutex
}

func newPingTracer() *pingTracer {
	return &pingTracer{
		pingID: map[string]struct{}{},
		Mutex:  &sync.Mutex{},
	}
}

func (t *pingTracer) add(id string) {
	t.Lock()
	defer t.Unlock()

	t.pingID[id] = struct{}{}
}

func (t *pingTracer) checkAndDelete(id string) bool {
	t.Lock()
	defer t.Unlock()

	_, ok := t.pingID[id]
	delete(t.pingID, id)
	return ok
}
