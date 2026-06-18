package mqrest

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

const attrMaxDepth = "maxdepth"
const attrDescr = "descr"
const attrTargq = "targq"
const attrMaxMsgl = "maxmsgl"
const attrMaxMsgLen = "maxmsglen"
const attrShare = "share"
const attrDefopts = "defopts"
const attrBothresh = "bothresh"
const attrBoqname = "boqname"
const attrUsage = "usage"
const attrTopstr = "topstr"
const attrTopicStr = "topicStr" // mqweb runCommandJSON name for TOPSTR
const attrReplace = "replace"
const attrTrptype = "trptype"
const attrMcaUser = "mcauser"
const attrSslPeer = "sslpeer"
const attrQmName = "qmname"

// queueDisplayParameters lists attributes safe for runCommandJSON DISPLAY qlocal
// on IBM MQ 9.4.x. Some keywords (e.g. maxmsglen, share, defopts) are rejected
// by mqweb with MQWB0120E even though they are valid on DEFINE.
var queueLocalDisplayParameters = []string{
	attrMaxDepth, attrDescr, "defpsist", "get", "put",
}

var queueAliasDisplayParameters = []string{
	attrTargq, "targtype", attrDescr,
}

const attrSslCiph = "sslciph"
const attrSslCauth = "sslcauth"

var queueRemoteDisplayParameters = []string{
	"rname", "rqmname", attrXmitq, attrDescr,
}

// queueNumericParameters are coerced to JSON numbers for runCommandJSON DEFINE.
var queueNumericParameters = map[string]struct{}{
	attrMaxDepth:  {},
	attrMaxMsgLen: {},
}

const (
	attrSharecnv = "sharecnv"
	attrMaxInst  = "maxinst"
	attrMaxInstc = "maxinstc"
	attrConname  = "conname"
	attrXmitq    = "xmitq"
)

var channelNumericParameters = map[string]struct{}{
	attrSharecnv: {},
	attrMaxMsgl:  {},
	attrMaxInst:  {},
	attrMaxInstc: {},
}

var channelSvrconnDisplayParameters = []string{
	attrDescr, attrTrptype, attrSharecnv, attrMaxMsgl, attrMcaUser, attrMaxInst, attrMaxInstc,
	attrSslCiph, attrSslCauth,
}

var channelSdrDisplayParameters = []string{
	attrDescr, attrTrptype, attrConname, attrXmitq, attrMaxMsgl, attrMcaUser, attrSslCiph,
}

var channelRcvrDisplayParameters = []string{
	attrDescr, attrTrptype, attrMaxMsgl, attrMcaUser, attrSslCiph,
}

// topicDisplayParameters lists attributes safe for DISPLAY topic on IBM MQ 9.4.x
// mqweb. pubscope/subscope are included for drift on 9.4; omit from this slice if
// your QM returns MQWB0120E (see docs/ATTRIBUTE_RECONCILIATION.md).
var topicDisplayParameters = []string{
	attrTopicStr, attrDescr, mqadmin.AttrKeyPub, mqadmin.AttrKeySub, "defpsist", "pubscope", "subscope",
}

var channelDisplayParameters = channelSvrconnDisplayParameters

func channelDisplayParametersForType(chlType mqadmin.ChannelType) []string {
	switch mqadmin.NormalizeChannelType(chlType) {
	case mqadmin.ChannelTypeSdr:
		return append([]string(nil), channelSdrDisplayParameters...)
	case mqadmin.ChannelTypeRcvr:
		return append([]string(nil), channelRcvrDisplayParameters...)
	default:
		return append([]string(nil), channelSvrconnDisplayParameters...)
	}
}

func defineTopicParameters(spec mqadmin.TopicSpec) map[string]any {
	params := defineObjectParameters(spec.Attributes, nil)
	mapTopicRESTParameters(params)
	return params
}

// mapTopicRESTParameters translates CRD/MQSC attribute names to mqweb JSON names.
func mapTopicRESTParameters(params map[string]any) {
	if v, ok := params[attrTopstr]; ok {
		params[attrTopicStr] = v
		delete(params, attrTopstr)
	}
	for _, key := range []string{mqadmin.AttrKeyPub, mqadmin.AttrKeySub} {
		if v, ok := params[key]; ok {
			params[key] = strings.ToUpper(fmt.Sprint(v))
		}
	}
}

func normalizeTopicAttributes(attrs map[string]string) {
	if v, ok := attrs[strings.ToLower(attrTopicStr)]; ok {
		attrs[attrTopstr] = v
	}
}

// normalizeQueueAttributes maps mqweb DISPLAY names to CRD/MQSC keys.
func normalizeQueueAttributes(attrs map[string]string, qType mqadmin.QueueType) {
	switch mqadmin.NormalizeQueueType(qType) {
	case mqadmin.QueueTypeAlias:
		if v, ok := attrs["target"]; ok && attrs[attrTargq] == "" {
			attrs[attrTargq] = v
		}
	case mqadmin.QueueTypeRemote:
		if v, ok := attrs["remotequeue"]; ok && attrs["rname"] == "" {
			attrs["rname"] = v
		}
		if v, ok := attrs["remotemanager"]; ok && attrs["rqmname"] == "" {
			attrs["rqmname"] = v
		}
		if v, ok := attrs["transmissionqueue"]; ok && attrs["xmitq"] == "" {
			attrs["xmitq"] = v
		}
	default:
	}
}

func defineChannelParameters(spec mqadmin.ChannelSpec) map[string]any {
	params := defineObjectParameters(spec.Attributes, channelNumericParameters)
	if spec.Type != "" {
		params["chltype"] = string(mqadmin.NormalizeChannelType(spec.Type))
	}
	return params
}

func validateChannelType(chType mqadmin.ChannelType) error {
	if mqadmin.ChannelTypeSupported(chType) {
		return nil
	}
	return &mqadmin.TerminalError{
		Reason:  "UnsupportedChannelType",
		Message: fmt.Sprintf("channel type %q is not supported", chType),
	}
}

func defineObjectParameters(
	attrs map[string]string,
	numericKeys map[string]struct{},
) map[string]any {
	params := map[string]any{attrReplace: mqscReplaceYes}
	for k, v := range attrs {
		key := strings.ToLower(k)
		if numericKeys != nil {
			if _, numeric := numericKeys[key]; numeric {
				if n, err := strconv.Atoi(v); err == nil {
					params[key] = n
					continue
				}
			}
		}
		params[key] = v
	}
	return params
}

func defineQueueParameters(spec mqadmin.QueueSpec) map[string]any {
	return defineObjectParameters(spec.Attributes, queueNumericParameters)
}

func queueQualifier(qType mqadmin.QueueType) string {
	switch mqadmin.NormalizeQueueType(qType) {
	case mqadmin.QueueTypeAlias:
		return qualifierQAlias
	case mqadmin.QueueTypeRemote:
		return qualifierQRemote
	default:
		return qualifierQLocal
	}
}

func queueDisplayParameters(qType mqadmin.QueueType) []string {
	switch mqadmin.NormalizeQueueType(qType) {
	case mqadmin.QueueTypeAlias:
		return append([]string(nil), queueAliasDisplayParameters...)
	case mqadmin.QueueTypeRemote:
		return append([]string(nil), queueRemoteDisplayParameters...)
	default:
		return append([]string(nil), queueLocalDisplayParameters...)
	}
}

func queueDisplayRequest(spec mqadmin.QueueSpec, responseParameters []string) runCommandJSONRequest {
	return runCommandJSONRequest{
		Type:               mqscType,
		Command:            mqscCommandDisplay,
		Qualifier:          queueQualifier(spec.Type),
		Name:               spec.Name,
		ResponseParameters: responseParameters,
	}
}

func channelDisplayRequest(name string, chlType mqadmin.ChannelType) runCommandJSONRequest {
	params := map[string]any{}
	if chlType != "" {
		params["chltype"] = string(mqadmin.NormalizeChannelType(chlType))
	}
	return runCommandJSONRequest{
		Type:               mqscType,
		Command:            mqscCommandDisplay,
		Qualifier:          qualifierChannel,
		Name:               name,
		Parameters:         params,
		ResponseParameters: channelDisplayParametersForType(chlType),
	}
}

// QueueDriftCheckKeys returns DISPLAY-safe queue attribute keys used for drift detection.
func QueueDriftCheckKeys(qType mqadmin.QueueType) []string {
	return queueDisplayParameters(qType)
}

// TopicDriftCheckKeys returns DISPLAY-safe topic attribute keys used for drift detection.
func TopicDriftCheckKeys() []string {
	return append([]string(nil), topicDisplayParameters...)
}

// ChannelDriftCheckKeys returns DISPLAY-safe channel attribute keys used for drift detection.
func ChannelDriftCheckKeys(chlType mqadmin.ChannelType) []string {
	return channelDisplayParametersForType(chlType)
}
