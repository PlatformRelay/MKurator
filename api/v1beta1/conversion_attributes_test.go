package v1beta1

import "testing"

func int32Ptr(v int32) *int32 { return &v }

func hasKey(m map[string]string, key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func TestFoldQueueAttributesToTyped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		spec   QueueSpec
		attrs  map[string]string
		assert func(t *testing.T, s QueueSpec, attrs map[string]string)
	}{
		{
			name:  "nil attrs is a no-op",
			spec:  QueueSpec{},
			attrs: nil,
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.MaxDepth != nil || attrs != nil {
					t.Fatalf("spec = %+v attrs = %v", s, attrs)
				}
			},
		},
		{
			name:  "folds maxdepth and drops the promoted key",
			spec:  QueueSpec{},
			attrs: map[string]string{"maxdepth": "5000", "custom": "keep"},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.MaxDepth == nil || *s.MaxDepth != 5000 {
					t.Fatalf("maxDepth = %v", s.MaxDepth)
				}
				if hasKey(attrs, "maxdepth") {
					t.Fatalf("promoted key not removed: %v", attrs)
				}
				if attrs["custom"] != "keep" {
					t.Fatalf("unrelated key lost: %v", attrs)
				}
			},
		},
		{
			name:  "case-insensitive key match",
			spec:  QueueSpec{},
			attrs: map[string]string{"MaxDepth": "42"},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.MaxDepth == nil || *s.MaxDepth != 42 {
					t.Fatalf("maxDepth = %v", s.MaxDepth)
				}
			},
		},
		{
			name:  "invalid maxdepth is not folded and key is retained",
			spec:  QueueSpec{},
			attrs: map[string]string{"maxdepth": "not-a-number"},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.MaxDepth != nil {
					t.Fatalf("maxDepth should stay unset on parse error: %v", s.MaxDepth)
				}
				if !hasKey(attrs, "maxdepth") {
					t.Fatalf("unparsed key should be retained: %v", attrs)
				}
			},
		},
		{
			name:  "typed field wins and conflicting attribute is dropped",
			spec:  QueueSpec{MaxDepth: int32Ptr(100)},
			attrs: map[string]string{"maxdepth": "9999"},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.MaxDepth == nil || *s.MaxDepth != 100 {
					t.Fatalf("typed field must win: %v", s.MaxDepth)
				}
				if hasKey(attrs, "maxdepth") {
					t.Fatalf("conflicting key should be dropped when typed field set: %v", attrs)
				}
			},
		},
		{
			name: "folds descr, defpsist, get, put string fields",
			spec: QueueSpec{},
			attrs: map[string]string{
				"descr":    "orders",
				"defpsist": "yes",
				"get":      "enabled",
				"put":      "disabled",
			},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.Description != "orders" || s.DefPersistence != QueueDefaultPersistenceYes {
					t.Fatalf("spec = %+v", s)
				}
				if s.Get != QueueAccessEnabledEnabled || s.Put != QueueAccessEnabledDisabled {
					t.Fatalf("spec = %+v", s)
				}
				if len(attrs) != 0 {
					t.Fatalf("all promoted keys should be removed: %v", attrs)
				}
			},
		},
		{
			name: "folds alternate alias/remote keys",
			spec: QueueSpec{},
			attrs: map[string]string{
				"target":            "APP.TARGET",
				"transmissionqueue": "SYSTEM.XMIT",
				"remotemanager":     "QM2",
			},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.TargetQueue != "APP.TARGET" || s.XmitQueue != "SYSTEM.XMIT" || s.RemoteQueueManager != "QM2" {
					t.Fatalf("spec = %+v", s)
				}
				if len(attrs) != 0 {
					t.Fatalf("all promoted keys should be removed: %v", attrs)
				}
			},
		},
		{
			name: "primary targq and xmitq keys fold",
			spec: QueueSpec{},
			attrs: map[string]string{
				"targq": "APP.ALIAS",
				"xmitq": "APP.XMIT",
			},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.TargetQueue != "APP.ALIAS" || s.XmitQueue != "APP.XMIT" {
					t.Fatalf("spec = %+v", s)
				}
			},
		},
		{
			name:  "unknown key is left untouched",
			spec:  QueueSpec{},
			attrs: map[string]string{"unknownkey": "value"},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if attrs["unknownkey"] != "value" {
					t.Fatalf("unknown key should be untouched: %v", attrs)
				}
			},
		},
		{
			name: "every typed queue field wins over its conflicting attribute",
			spec: QueueSpec{
				MaxDepth:           int32Ptr(1),
				Description:        "typed",
				DefPersistence:     QueueDefaultPersistenceYes,
				Get:                QueueAccessEnabledEnabled,
				Put:                QueueAccessEnabledEnabled,
				TargetQueue:        "T",
				XmitQueue:          "X",
				RemoteQueueManager: "R",
			},
			attrs: map[string]string{
				"maxdepth": "9", "descr": "a", "defpsist": "no", "get": "disabled",
				"put": "disabled", "targq": "a", "xmitq": "a", "rqmname": "a",
			},
			assert: func(t *testing.T, s QueueSpec, attrs map[string]string) {
				t.Helper()
				if s.Description != "typed" || *s.MaxDepth != 1 {
					t.Fatalf("typed fields must be preserved: %+v", s)
				}
				if len(attrs) != 0 {
					t.Fatalf("all conflicting keys should be dropped when typed set: %v", attrs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spec := tt.spec
			FoldQueueAttributesToTyped(&spec, tt.attrs)
			tt.assert(t, spec, tt.attrs)
		})
	}
}

func TestFoldTopicAttributesToTyped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		spec   TopicSpec
		attrs  map[string]string
		assert func(t *testing.T, s TopicSpec, attrs map[string]string)
	}{
		{
			name:  "nil attrs is a no-op",
			spec:  TopicSpec{},
			attrs: nil,
			assert: func(t *testing.T, s TopicSpec, attrs map[string]string) {
				t.Helper()
				if s.TopicString != "" || attrs != nil {
					t.Fatalf("spec = %+v attrs = %v", s, attrs)
				}
			},
		},
		{
			name: "folds all promoted topic keys and clears the map",
			spec: TopicSpec{},
			attrs: map[string]string{
				"topstr":   "/prices",
				"descr":    "price feed",
				"pub":      "enabled",
				"sub":      "disabled",
				"defpsist": "no",
				"pubscope": "all",
				"subscope": "qmgr",
			},
			assert: func(t *testing.T, s TopicSpec, attrs map[string]string) {
				t.Helper()
				if s.TopicString != "/prices" || s.Description != "price feed" {
					t.Fatalf("spec = %+v", s)
				}
				if s.Publish != TopicAccessEnabledEnabled || s.Subscribe != TopicAccessEnabledDisabled {
					t.Fatalf("spec = %+v", s)
				}
				if s.DefPersistence != QueueDefaultPersistenceNo {
					t.Fatalf("spec = %+v", s)
				}
				if s.PublishScope != "all" || s.SubscribeScope != "qmgr" {
					t.Fatalf("spec = %+v", s)
				}
				if len(attrs) != 0 {
					t.Fatalf("all promoted keys should be removed: %v", attrs)
				}
			},
		},
		{
			name:  "topicstr alias normalizes to topstr",
			spec:  TopicSpec{},
			attrs: map[string]string{"TopicStr": "/orders"},
			assert: func(t *testing.T, s TopicSpec, attrs map[string]string) {
				t.Helper()
				if s.TopicString != "/orders" {
					t.Fatalf("topicString = %q", s.TopicString)
				}
				if hasKey(attrs, "TopicStr") {
					t.Fatalf("promoted alias key should be removed: %v", attrs)
				}
			},
		},
		{
			name:  "typed topic string wins over attribute",
			spec:  TopicSpec{TopicString: "/typed"},
			attrs: map[string]string{"topstr": "/attr"},
			assert: func(t *testing.T, s TopicSpec, attrs map[string]string) {
				t.Helper()
				if s.TopicString != "/typed" {
					t.Fatalf("typed field must win: %q", s.TopicString)
				}
				if hasKey(attrs, "topstr") {
					t.Fatalf("conflicting key should be dropped: %v", attrs)
				}
			},
		},
		{
			name: "every typed topic field wins over its conflicting attribute",
			spec: TopicSpec{
				TopicString:    "/typed",
				Description:    "typed",
				Publish:        TopicAccessEnabledEnabled,
				Subscribe:      TopicAccessEnabledEnabled,
				DefPersistence: QueueDefaultPersistenceYes,
				PublishScope:   "all",
				SubscribeScope: "all",
			},
			attrs: map[string]string{
				"topstr": "/a", "descr": "a", "pub": "disabled", "sub": "disabled",
				"defpsist": "no", "pubscope": "qmgr", "subscope": "qmgr",
			},
			assert: func(t *testing.T, s TopicSpec, attrs map[string]string) {
				t.Helper()
				if s.TopicString != "/typed" || s.Description != "typed" {
					t.Fatalf("typed fields must be preserved: %+v", s)
				}
				if len(attrs) != 0 {
					t.Fatalf("all conflicting keys should be dropped when typed set: %v", attrs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spec := tt.spec
			FoldTopicAttributesToTyped(&spec, tt.attrs)
			tt.assert(t, spec, tt.attrs)
		})
	}
}

func TestFoldChannelAttributesToTyped(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		spec   ChannelSpec
		attrs  map[string]string
		assert func(t *testing.T, s ChannelSpec, attrs map[string]string)
	}{
		{
			name:  "nil attrs is a no-op",
			spec:  ChannelSpec{},
			attrs: nil,
			assert: func(t *testing.T, s ChannelSpec, attrs map[string]string) {
				t.Helper()
				if s.Description != "" || attrs != nil {
					t.Fatalf("spec = %+v attrs = %v", s, attrs)
				}
			},
		},
		{
			name: "folds all promoted channel keys and clears the map",
			spec: ChannelSpec{},
			attrs: map[string]string{
				"descr":    "svrconn",
				"maxmsgl":  "4194304",
				"trptype":  "tcp",
				"sharecnv": "10",
				"mcauser":  "mqm",
				"maxinst":  "50",
				"maxinstc": "5",
				"sslciph":  "TLS_RSA",
				"sslcauth": "required",
				"conname":  "host(1414)",
				"xmitq":    "SYSTEM.XMIT",
			},
			assert: func(t *testing.T, s ChannelSpec, attrs map[string]string) {
				t.Helper()
				if s.Description != "svrconn" || s.McaUser != "mqm" {
					t.Fatalf("spec = %+v", s)
				}
				if s.MaxMsgLength == nil || *s.MaxMsgLength != 4194304 {
					t.Fatalf("maxMsgLength = %v", s.MaxMsgLength)
				}
				if s.TransportType != ChannelTransportTypeTCP {
					t.Fatalf("transportType = %q", s.TransportType)
				}
				if s.ShareConv == nil || *s.ShareConv != 10 {
					t.Fatalf("shareConv = %v", s.ShareConv)
				}
				if s.MaxInstances == nil || *s.MaxInstances != 50 {
					t.Fatalf("maxInstances = %v", s.MaxInstances)
				}
				if s.MaxInstancesClient == nil || *s.MaxInstancesClient != 5 {
					t.Fatalf("maxInstancesClient = %v", s.MaxInstancesClient)
				}
				if s.SslCipherSpec != "TLS_RSA" || s.SslClientAuth != ChannelSslClientAuthRequired {
					t.Fatalf("spec = %+v", s)
				}
				if s.ConnName != "host(1414)" || s.XmitQueue != "SYSTEM.XMIT" {
					t.Fatalf("spec = %+v", s)
				}
				if len(attrs) != 0 {
					t.Fatalf("all promoted keys should be removed: %v", attrs)
				}
			},
		},
		{
			name:  "case-insensitive channel key match",
			spec:  ChannelSpec{},
			attrs: map[string]string{"MAXMSGL": "1024"},
			assert: func(t *testing.T, s ChannelSpec, attrs map[string]string) {
				t.Helper()
				if s.MaxMsgLength == nil || *s.MaxMsgLength != 1024 {
					t.Fatalf("maxMsgLength = %v", s.MaxMsgLength)
				}
			},
		},
		{
			name:  "invalid numeric channel value is retained unfolded",
			spec:  ChannelSpec{},
			attrs: map[string]string{"sharecnv": "abc"},
			assert: func(t *testing.T, s ChannelSpec, attrs map[string]string) {
				t.Helper()
				if s.ShareConv != nil {
					t.Fatalf("shareConv should stay unset on parse error: %v", s.ShareConv)
				}
				if !hasKey(attrs, "sharecnv") {
					t.Fatalf("unparsed key should be retained: %v", attrs)
				}
			},
		},
		{
			name:  "typed channel field wins over conflicting attribute",
			spec:  ChannelSpec{Description: "typed"},
			attrs: map[string]string{"descr": "attr"},
			assert: func(t *testing.T, s ChannelSpec, attrs map[string]string) {
				t.Helper()
				if s.Description != "typed" {
					t.Fatalf("typed field must win: %q", s.Description)
				}
				if hasKey(attrs, "descr") {
					t.Fatalf("conflicting key should be dropped: %v", attrs)
				}
			},
		},
		{
			name: "every typed channel field wins over its conflicting attribute",
			spec: ChannelSpec{
				Description:        "typed",
				MaxMsgLength:       int32Ptr(1),
				TransportType:      ChannelTransportTypeTCP,
				ShareConv:          int32Ptr(1),
				McaUser:            "u",
				MaxInstances:       int32Ptr(1),
				MaxInstancesClient: int32Ptr(1),
				SslCipherSpec:      "c",
				SslClientAuth:      ChannelSslClientAuthRequired,
				ConnName:           "n",
				XmitQueue:          "x",
			},
			attrs: map[string]string{
				"descr": "a", "maxmsgl": "9", "trptype": "lu62", "sharecnv": "9",
				"mcauser": "a", "maxinst": "9", "maxinstc": "9", "sslciph": "a",
				"sslcauth": "optional", "conname": "a", "xmitq": "a",
			},
			assert: func(t *testing.T, s ChannelSpec, attrs map[string]string) {
				t.Helper()
				if s.Description != "typed" || s.McaUser != "u" {
					t.Fatalf("typed fields must be preserved: %+v", s)
				}
				if len(attrs) != 0 {
					t.Fatalf("all conflicting keys should be dropped when typed set: %v", attrs)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			spec := tt.spec
			FoldChannelAttributesToTyped(&spec, tt.attrs)
			tt.assert(t, spec, tt.attrs)
		})
	}
}

func TestCloneStringMap(t *testing.T) {
	t.Parallel()

	if got := CloneStringMap(nil); got != nil {
		t.Fatalf("nil map should clone to nil, got %v", got)
	}
	if got := CloneStringMap(map[string]string{}); got != nil {
		t.Fatalf("empty map should clone to nil, got %v", got)
	}

	src := map[string]string{"a": "1", "b": "2"}
	clone := CloneStringMap(src)
	if len(clone) != 2 || clone["a"] != "1" || clone["b"] != "2" {
		t.Fatalf("clone = %v", clone)
	}
	// Mutating the clone must not affect the source.
	clone["a"] = "mutated"
	if src["a"] != "1" {
		t.Fatalf("source mutated through clone: %v", src)
	}
}
