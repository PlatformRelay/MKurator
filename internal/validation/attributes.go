package validation

import (
	"fmt"

	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

const unknownAttrWarningFmt = "attribute key %q is not in the drift-check allow-list; " +
	"it may still be applied by MQ but is ignored for drift detection"

var (
	queueLocalKnownAttrs = keys(
		attrKeyMaxDepth, attrKeyDescr, "defpsist", "get", "put",
		"maxmsglen", "share", "defopts", "bothresh", "boqname", "usage",
		"trigtype", "trigdata", "trigmpri", "trigint", "trigdpth",
		"cluster", "clusnl",
	)
	queueAliasKnownAttrs  = keys(attrKeyTargQ, "targtype", attrKeyDescr, attrKeyTarget)
	queueRemoteKnownAttrs = keys(
		"rname", attrKeyRQMName, attrKeyXmitQ, attrKeyDescr,
		"remotequeue", attrKeyRemoteMgr, attrKeyXmitQLong,
	)
	topicKnownAttrs = keys(
		attrKeyTopStr, attrKeyTopicStr, attrKeyDescr, mqadmin.AttrKeyPub, mqadmin.AttrKeySub, "defpsist",
		"pubscope", "subscope", "toptype", "cluster",
	)
	channelKnownAttrs = keys(
		attrKeyDescr, "trptype", "sharecnv", attrKeyMaxMsgL, "mcauser", "maxinst", "maxinstc",
		"sslciph", "sslcauth", attrKeyConName, attrKeyXmitQ,
	)
)

func keys(names ...string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, n := range names {
		out[n] = struct{}{}
	}
	return out
}

func unknownQueueAttributeWarnings(qType mqadmin.QueueType, attrs map[string]string) []string {
	known := queueLocalKnownAttrs
	switch mqadmin.NormalizeQueueType(qType) {
	case mqadmin.QueueTypeAlias:
		known = queueAliasKnownAttrs
	case mqadmin.QueueTypeRemote:
		known = queueRemoteKnownAttrs
	}
	return unknownAttributeWarnings(attrs, known)
}

func unknownTopicAttributeWarnings(attrs map[string]string) []string {
	return unknownAttributeWarnings(attrs, topicKnownAttrs)
}

func unknownChannelAttributeWarnings(attrs map[string]string) []string {
	return unknownAttributeWarnings(attrs, channelKnownAttrs)
}

func unknownAttributeWarnings(attrs map[string]string, known map[string]struct{}) []string {
	if len(attrs) == 0 {
		return nil
	}
	warnings := make([]string, 0, len(attrs))
	for k := range attrs {
		key := mqadmin.NormalizeAttrKey(k)
		if _, ok := known[key]; ok {
			continue
		}
		if _, ok := known[k]; ok {
			continue
		}
		warnings = append(warnings, fmt.Sprintf(unknownAttrWarningFmt, k))
	}
	return warnings
}
