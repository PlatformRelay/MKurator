package validation

// Shared IBM MQ attribute key literals (allow-list + deprecated maps).
const (
	attrKeyMaxDepth  = "maxdepth"
	attrKeyDescr     = "descr"
	attrKeyTargQ     = "targq"
	attrKeyTarget    = "target"
	attrKeyXmitQ     = "xmitq"
	attrKeyXmitQLong = "transmissionqueue"
	attrKeyRQMName   = "rqmname"
	attrKeyRemoteMgr = "remotemanager"
	attrKeyTopStr    = "topstr"
	attrKeyTopicStr  = "topicstr"
	attrKeyPub       = "pub"
	attrKeyMaxMsgL   = "maxmsgl"
	attrKeyConName   = "conname"
)

const (
	specPathMaxDepth           = "spec.maxDepth"
	specPathDescription        = "spec.description"
	specPathTargetQueue        = "spec.targetQueue"
	specPathXmitQueue          = "spec.xmitQueue"
	specPathRemoteQueueManager = "spec.remoteQueueManager"
	specPathTopicString        = "spec.topicString"
	specPathPublish            = "spec.publish"
	specPathMaxMsgLength       = "spec.maxMsgLength"
	specPathConnName           = "spec.connName"
)
