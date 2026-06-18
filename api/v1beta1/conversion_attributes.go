package v1beta1

import (
	"strconv"
	"strings"
)

const (
	attrKeyDescr    = "descr"
	attrKeyDefpsist = "defpsist"
	attrKeyXmitq    = "xmitq"
	attrKeyTopicstr = "topicstr"
	attrKeyTopstr   = "topstr"
)

type attrFold struct {
	keys  []string
	apply func(spec *QueueSpec, value string) bool
}

// FoldQueueAttributesToTyped copies promoted attribute keys into typed QueueSpec fields when
// the typed field is unset. Promoted keys are removed from attrs; when both map and typed
// values exist, typed fields already on spec win.
func FoldQueueAttributesToTyped(spec *QueueSpec, attrs map[string]string) {
	if len(attrs) == 0 {
		return
	}
	folds := []attrFold{
		{keys: []string{"maxdepth"}, apply: func(s *QueueSpec, v string) bool {
			if s.MaxDepth != nil {
				return false
			}
			n, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return false
			}
			depth := int32(n)
			s.MaxDepth = &depth
			return true
		}},
		{keys: []string{attrKeyDescr}, apply: func(s *QueueSpec, v string) bool {
			if s.Description != "" {
				return false
			}
			s.Description = v
			return true
		}},
		{keys: []string{attrKeyDefpsist}, apply: func(s *QueueSpec, v string) bool {
			if s.DefPersistence != "" {
				return false
			}
			s.DefPersistence = QueueDefaultPersistence(v)
			return true
		}},
		{keys: []string{"get"}, apply: func(s *QueueSpec, v string) bool {
			if s.Get != "" {
				return false
			}
			s.Get = QueueAccessEnabled(v)
			return true
		}},
		{keys: []string{"put"}, apply: func(s *QueueSpec, v string) bool {
			if s.Put != "" {
				return false
			}
			s.Put = QueueAccessEnabled(v)
			return true
		}},
		{keys: []string{"targq", "target"}, apply: func(s *QueueSpec, v string) bool {
			if s.TargetQueue != "" {
				return false
			}
			s.TargetQueue = v
			return true
		}},
		{keys: []string{attrKeyXmitq, "transmissionqueue"}, apply: func(s *QueueSpec, v string) bool {
			if s.XmitQueue != "" {
				return false
			}
			s.XmitQueue = v
			return true
		}},
		{keys: []string{"rqmname", "remotemanager"}, apply: func(s *QueueSpec, v string) bool {
			if s.RemoteQueueManager != "" {
				return false
			}
			s.RemoteQueueManager = v
			return true
		}},
	}
	foldAttributes(spec, attrs, folds, func(f attrFold, v string) bool { return f.apply(spec, v) })
}

type topicAttrFold struct {
	keys  []string
	apply func(spec *TopicSpec, value string) bool
}

// FoldTopicAttributesToTyped copies promoted attribute keys into typed TopicSpec fields when unset.
func FoldTopicAttributesToTyped(spec *TopicSpec, attrs map[string]string) {
	if len(attrs) == 0 {
		return
	}
	folds := []topicAttrFold{
		{keys: []string{attrKeyTopstr, attrKeyTopicstr}, apply: func(s *TopicSpec, v string) bool {
			if s.TopicString != "" {
				return false
			}
			s.TopicString = v
			return true
		}},
		{keys: []string{attrKeyDescr}, apply: func(s *TopicSpec, v string) bool {
			if s.Description != "" {
				return false
			}
			s.Description = v
			return true
		}},
		{keys: []string{"pub"}, apply: func(s *TopicSpec, v string) bool {
			if s.Publish != "" {
				return false
			}
			s.Publish = TopicAccessEnabled(v)
			return true
		}},
		{keys: []string{"sub"}, apply: func(s *TopicSpec, v string) bool {
			if s.Subscribe != "" {
				return false
			}
			s.Subscribe = TopicAccessEnabled(v)
			return true
		}},
		{keys: []string{attrKeyDefpsist}, apply: func(s *TopicSpec, v string) bool {
			if s.DefPersistence != "" {
				return false
			}
			s.DefPersistence = QueueDefaultPersistence(v)
			return true
		}},
		{keys: []string{"pubscope"}, apply: func(s *TopicSpec, v string) bool {
			if s.PublishScope != "" {
				return false
			}
			s.PublishScope = v
			return true
		}},
		{keys: []string{"subscope"}, apply: func(s *TopicSpec, v string) bool {
			if s.SubscribeScope != "" {
				return false
			}
			s.SubscribeScope = v
			return true
		}},
	}
	for key, value := range attrs {
		norm := normalizeTopicAttrKey(key)
		for _, fold := range folds {
			if !attrKeyMatches(norm, fold.keys) {
				continue
			}
			if fold.apply(spec, value) || typedTopicFieldSet(spec, fold.keys) {
				delete(attrs, key)
			}
			break
		}
	}
}

type channelAttrFold struct {
	keys  []string
	apply func(spec *ChannelSpec, value string) bool
}

// FoldChannelAttributesToTyped copies promoted attribute keys into typed ChannelSpec fields when unset.
//
//nolint:gocyclo // table-driven folds over many MQSC keys.
func FoldChannelAttributesToTyped(spec *ChannelSpec, attrs map[string]string) {
	if len(attrs) == 0 {
		return
	}
	folds := []channelAttrFold{
		{keys: []string{attrKeyDescr}, apply: func(s *ChannelSpec, v string) bool {
			if s.Description != "" {
				return false
			}
			s.Description = v
			return true
		}},
		{keys: []string{"maxmsgl"}, apply: func(s *ChannelSpec, v string) bool {
			if s.MaxMsgLength != nil {
				return false
			}
			n, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return false
			}
			val := int32(n)
			s.MaxMsgLength = &val
			return true
		}},
		{keys: []string{"trptype"}, apply: func(s *ChannelSpec, v string) bool {
			if s.TransportType != "" {
				return false
			}
			s.TransportType = ChannelTransportType(v)
			return true
		}},
		{keys: []string{"sharecnv"}, apply: func(s *ChannelSpec, v string) bool {
			if s.ShareConv != nil {
				return false
			}
			n, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return false
			}
			val := int32(n)
			s.ShareConv = &val
			return true
		}},
		{keys: []string{"mcauser"}, apply: func(s *ChannelSpec, v string) bool {
			if s.McaUser != "" {
				return false
			}
			s.McaUser = v
			return true
		}},
		{keys: []string{"maxinst"}, apply: func(s *ChannelSpec, v string) bool {
			if s.MaxInstances != nil {
				return false
			}
			n, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return false
			}
			val := int32(n)
			s.MaxInstances = &val
			return true
		}},
		{keys: []string{"maxinstc"}, apply: func(s *ChannelSpec, v string) bool {
			if s.MaxInstancesClient != nil {
				return false
			}
			n, err := strconv.ParseInt(v, 10, 32)
			if err != nil {
				return false
			}
			val := int32(n)
			s.MaxInstancesClient = &val
			return true
		}},
		{keys: []string{"sslciph"}, apply: func(s *ChannelSpec, v string) bool {
			if s.SslCipherSpec != "" {
				return false
			}
			s.SslCipherSpec = v
			return true
		}},
		{keys: []string{"sslcauth"}, apply: func(s *ChannelSpec, v string) bool {
			if s.SslClientAuth != "" {
				return false
			}
			s.SslClientAuth = ChannelSslClientAuth(v)
			return true
		}},
		{keys: []string{"conname"}, apply: func(s *ChannelSpec, v string) bool {
			if s.ConnName != "" {
				return false
			}
			s.ConnName = v
			return true
		}},
		{keys: []string{attrKeyXmitq}, apply: func(s *ChannelSpec, v string) bool {
			if s.XmitQueue != "" {
				return false
			}
			s.XmitQueue = v
			return true
		}},
	}
	for key, value := range attrs {
		norm := strings.ToLower(key)
		for _, fold := range folds {
			if !attrKeyMatches(norm, fold.keys) {
				continue
			}
			if fold.apply(spec, value) || typedChannelFieldSet(spec, fold.keys) {
				delete(attrs, key)
			}
			break
		}
	}
}

func foldAttributes(spec *QueueSpec, attrs map[string]string, folds []attrFold, apply func(attrFold, string) bool) {
	for key, value := range attrs {
		norm := strings.ToLower(key)
		for _, fold := range folds {
			if !attrKeyMatches(norm, fold.keys) {
				continue
			}
			if apply(fold, value) || typedQueueFieldSet(spec, fold.keys) {
				delete(attrs, key)
			}
			break
		}
	}
}

func attrKeyMatches(norm string, keys []string) bool {
	for _, k := range keys {
		if norm == k {
			return true
		}
	}
	return false
}

func normalizeTopicAttrKey(key string) string {
	norm := strings.ToLower(key)
	if norm == attrKeyTopicstr {
		return attrKeyTopstr
	}
	return norm
}

func typedQueueFieldSet(spec *QueueSpec, keys []string) bool {
	for _, k := range keys {
		switch k {
		case "maxdepth":
			if spec.MaxDepth != nil {
				return true
			}
		case attrKeyDescr:
			if spec.Description != "" {
				return true
			}
		case attrKeyDefpsist:
			if spec.DefPersistence != "" {
				return true
			}
		case "get":
			if spec.Get != "" {
				return true
			}
		case "put":
			if spec.Put != "" {
				return true
			}
		case "targq", "target":
			if spec.TargetQueue != "" {
				return true
			}
		case attrKeyXmitq, "transmissionqueue":
			if spec.XmitQueue != "" {
				return true
			}
		case "rqmname", "remotemanager":
			if spec.RemoteQueueManager != "" {
				return true
			}
		}
	}
	return false
}

func typedTopicFieldSet(spec *TopicSpec, keys []string) bool {
	for _, k := range keys {
		switch k {
		case "topstr", "topicstr":
			if spec.TopicString != "" {
				return true
			}
		case attrKeyDescr:
			if spec.Description != "" {
				return true
			}
		case "pub":
			if spec.Publish != "" {
				return true
			}
		case "sub":
			if spec.Subscribe != "" {
				return true
			}
		case attrKeyDefpsist:
			if spec.DefPersistence != "" {
				return true
			}
		case "pubscope":
			if spec.PublishScope != "" {
				return true
			}
		case "subscope":
			if spec.SubscribeScope != "" {
				return true
			}
		}
	}
	return false
}

//nolint:gocyclo // mirrors promoted-key table for ChannelSpec.
func typedChannelFieldSet(spec *ChannelSpec, keys []string) bool {
	for _, k := range keys {
		switch k {
		case attrKeyDescr:
			if spec.Description != "" {
				return true
			}
		case "maxmsgl":
			if spec.MaxMsgLength != nil {
				return true
			}
		case "trptype":
			if spec.TransportType != "" {
				return true
			}
		case "sharecnv":
			if spec.ShareConv != nil {
				return true
			}
		case "mcauser":
			if spec.McaUser != "" {
				return true
			}
		case "maxinst":
			if spec.MaxInstances != nil {
				return true
			}
		case "maxinstc":
			if spec.MaxInstancesClient != nil {
				return true
			}
		case "sslciph":
			if spec.SslCipherSpec != "" {
				return true
			}
		case "sslcauth":
			if spec.SslClientAuth != "" {
				return true
			}
		case "conname":
			if spec.ConnName != "" {
				return true
			}
		case attrKeyXmitq:
			if spec.XmitQueue != "" {
				return true
			}
		}
	}
	return false
}

// CloneStringMap returns a shallow copy of attrs, or nil when empty.
func CloneStringMap(attrs map[string]string) map[string]string {
	if len(attrs) == 0 {
		return nil
	}
	out := make(map[string]string, len(attrs))
	for k, v := range attrs {
		out[k] = v
	}
	return out
}
