package validation

import (
	"fmt"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

const deprecatedAttrWarningFmt = `spec.attributes[%q] is deprecated; use %s`

var (
	deprecatedQueueAttributeFields = map[string]string{
		attrKeyMaxDepth:  specPathMaxDepth,
		attrKeyDescr:     specPathDescription,
		"defpsist":       "spec.defPersistence",
		"get":            "spec.get",
		"put":            "spec.put",
		attrKeyTargQ:     specPathTargetQueue,
		attrKeyTarget:    specPathTargetQueue,
		attrKeyXmitQ:     specPathXmitQueue,
		attrKeyXmitQLong: specPathXmitQueue,
		attrKeyRQMName:   specPathRemoteQueueManager,
		attrKeyRemoteMgr: specPathRemoteQueueManager,
	}
	deprecatedTopicAttributeFields = map[string]string{
		attrKeyTopStr:   specPathTopicString,
		attrKeyTopicStr: specPathTopicString,
		attrKeyDescr:    specPathDescription,
		attrKeyPub:      specPathPublish,
		"sub":           "spec.subscribe",
		"defpsist":      "spec.defPersistence",
		"pubscope":      "spec.publishScope",
		"subscope":      "spec.subscribeScope",
	}
	deprecatedChannelAttributeFields = map[string]string{
		attrKeyDescr:   specPathDescription,
		attrKeyMaxMsgL: specPathMaxMsgLength,
		"trptype":      "spec.transportType",
		"sharecnv":     "spec.shareConv",
		"mcauser":      "spec.mcaUser",
		"maxinst":      "spec.maxInstances",
		"maxinstc":     "spec.maxInstancesClient",
		"sslciph":      "spec.sslCipherSpec",
		"sslcauth":     "spec.sslClientAuth",
		attrKeyConName: specPathConnName,
		attrKeyXmitQ:   specPathXmitQueue,
	}
)

func deprecatedQueueAttributeWarnings(attrs map[string]string) []string {
	return deprecatedAttributeWarnings(attrs, deprecatedQueueAttributeFields)
}

func deprecatedTopicAttributeWarnings(attrs map[string]string) []string {
	return deprecatedAttributeWarnings(attrs, deprecatedTopicAttributeFields)
}

func deprecatedChannelAttributeWarnings(attrs map[string]string) []string {
	return deprecatedAttributeWarnings(attrs, deprecatedChannelAttributeFields)
}

func deprecatedAttributeWarnings(attrs map[string]string, replacements map[string]string) []string {
	if len(attrs) == 0 {
		return nil
	}
	warnings := make([]string, 0, len(attrs))
	for key := range attrs {
		replacement, ok := replacements[mqadmin.NormalizeAttrKey(key)]
		if !ok {
			continue
		}
		warnings = append(warnings, fmt.Sprintf(deprecatedAttrWarningFmt, key, replacement))
	}
	return warnings
}
