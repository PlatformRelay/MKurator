package mqrest

import (
	"strings"
	"testing"

	"github.com/conduit-ops/mkurator/internal/mqadmin"
)

func TestDefineQueueParameters(t *testing.T) {
	t.Parallel()
	params := defineQueueParameters(mqadmin.QueueSpec{
		Name: "APP.ORDERS",
		Attributes: map[string]string{
			attrMaxDepth: "5000",
			"descr":      "orders",
		},
	})
	if params["replace"] != "yes" {
		t.Fatalf("replace = %v", params["replace"])
	}
	if params[attrMaxDepth] != 5000 {
		t.Fatalf("maxdepth should be int 5000, got %T(%v)", params[attrMaxDepth], params[attrMaxDepth])
	}
	if params["descr"] != "orders" {
		t.Fatalf("descr = %v", params["descr"])
	}
}

func TestQueueDisplayParametersExcludeMaxmsglen(t *testing.T) {
	t.Parallel()
	for _, p := range queueLocalDisplayParameters {
		if p == attrMaxMsgLen {
			t.Fatal("maxmsglen must not be in display parameters for mqweb 9.4")
		}
	}
}

func TestQueueQualifier(t *testing.T) {
	t.Parallel()
	if queueQualifier(mqadmin.QueueTypeLocal) != "qlocal" {
		t.Fatal("local")
	}
	if queueQualifier(mqadmin.QueueTypeAlias) != "qalias" {
		t.Fatal("alias")
	}
	if queueQualifier(mqadmin.QueueTypeRemote) != "qremote" {
		t.Fatal("remote")
	}
}

func TestDefineChannelParameters(t *testing.T) {
	t.Parallel()
	params := defineChannelParameters(mqadmin.ChannelSpec{
		Name: "ORDERS.APP",
		Type: mqadmin.ChannelTypeSvrconn,
		Attributes: map[string]string{
			"maxmsgl": "4194304",
			"trptype": "tcp",
		},
	})
	if params["chltype"] != "svrconn" {
		t.Fatalf("chltype = %v", params["chltype"])
	}
	if params["maxmsgl"] != 4194304 {
		t.Fatalf("maxmsgl should be int, got %T(%v)", params["maxmsgl"], params["maxmsgl"])
	}
}

func TestChannelDisplayParametersIncludeConnectionLimits(t *testing.T) {
	t.Parallel()
	want := map[string]struct{}{"maxinst": {}, "maxinstc": {}, "sslciph": {}, "sslcauth": {}}
	for k := range want {
		found := false
		for _, p := range channelSvrconnDisplayParameters {
			if p == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%q missing from channelSvrconnDisplayParameters", k)
		}
	}
}

func TestChannelSdrDisplayParametersIncludeConnectionAttrs(t *testing.T) {
	t.Parallel()
	for _, k := range []string{attrConname, attrXmitq} {
		found := false
		for _, p := range channelSdrDisplayParameters {
			if p == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%q missing from channelSdrDisplayParameters", k)
		}
	}
}

func TestChannelRcvrDisplayParametersOmitConnectionAttrs(t *testing.T) {
	t.Parallel()
	for _, k := range []string{attrDescr, attrTrptype, attrMaxMsgl, attrMcaUser, attrSslCiph} {
		found := false
		for _, p := range channelRcvrDisplayParameters {
			if p == k {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%q missing from channelRcvrDisplayParameters", k)
		}
	}
	for _, k := range []string{attrConname, attrXmitq} {
		for _, p := range channelRcvrDisplayParameters {
			if p == k {
				t.Fatalf("%q must not be in channelRcvrDisplayParameters", k)
			}
		}
	}
}

func TestDefineChannelParametersRcvr(t *testing.T) {
	t.Parallel()
	params := defineChannelParameters(mqadmin.ChannelSpec{
		Name: "QM2.FROM.QM1",
		Type: mqadmin.ChannelTypeRcvr,
		Attributes: map[string]string{
			"trptype": "tcp",
			"descr":   "inbound",
		},
	})
	if params["chltype"] != "rcvr" {
		t.Fatalf("chltype = %v", params["chltype"])
	}
	if params["trptype"] != "tcp" {
		t.Fatalf("trptype = %v", params["trptype"])
	}
}

func TestDefineChannelParametersSdr(t *testing.T) {
	t.Parallel()
	params := defineChannelParameters(mqadmin.ChannelSpec{
		Name: "QM1.TO.QM2",
		Type: mqadmin.ChannelTypeSdr,
		Attributes: map[string]string{
			"conname": "qm2.example.com(1414)",
			"xmitq":   "SYSTEM.DEFAULT.XMIT.QUEUE",
			"trptype": "tcp",
		},
	})
	if params["chltype"] != "sdr" {
		t.Fatalf("chltype = %v", params["chltype"])
	}
	if params["conname"] != "qm2.example.com(1414)" {
		t.Fatalf("conname = %v", params["conname"])
	}
}

func TestValidateChannelType(t *testing.T) {
	t.Parallel()
	if err := validateChannelType(mqadmin.ChannelTypeSvrconn); err != nil {
		t.Fatalf("svrconn: %v", err)
	}
	if err := validateChannelType(mqadmin.ChannelTypeSdr); err != nil {
		t.Fatalf("sdr: %v", err)
	}
	if err := validateChannelType(mqadmin.ChannelTypeRcvr); err != nil {
		t.Fatalf("rcvr: %v", err)
	}
	if err := validateChannelType(mqadmin.ChannelType("clusrcv")); err == nil {
		t.Fatal("expected error for clusrcv")
	}
}

func TestQueueLocalDisplayParametersOmitExtendedAttrsOn94(t *testing.T) {
	t.Parallel()
	for _, p := range []string{attrShare, attrDefopts, attrBothresh, attrBoqname, attrUsage} {
		for _, q := range queueLocalDisplayParameters {
			if q == p {
				t.Fatalf("%q must not be in queueLocalDisplayParameters on mqweb 9.4 (MQWB0120E)", p)
			}
		}
	}
}

func TestDriftCheckKeyExports(t *testing.T) {
	t.Parallel()
	if len(QueueDriftCheckKeys(mqadmin.QueueTypeLocal)) == 0 {
		t.Fatal("expected local queue drift keys")
	}
	if len(TopicDriftCheckKeys()) == 0 {
		t.Fatal("expected topic drift keys")
	}
	if len(ChannelDriftCheckKeys(mqadmin.ChannelTypeSvrconn)) == 0 {
		t.Fatal("expected channel drift keys")
	}
	if len(ChannelDriftCheckKeys(mqadmin.ChannelTypeSdr)) == 0 {
		t.Fatal("expected sdr channel drift keys")
	}
	if len(ChannelDriftCheckKeys(mqadmin.ChannelTypeRcvr)) == 0 {
		t.Fatal("expected rcvr channel drift keys")
	}
}

func TestTopicDisplayParametersIncludeScope(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"pubscope", "subscope"} {
		found := false
		for _, q := range topicDisplayParameters {
			if q == p {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%q missing from topicDisplayParameters", p)
		}
	}
}

func TestDefineTopicParameters(t *testing.T) {
	t.Parallel()
	params := defineTopicParameters(mqadmin.TopicSpec{
		Name: "RETAIL.ORDERS",
		Attributes: map[string]string{
			"topstr": "retail/orders",
		},
	})
	if params["replace"] != "yes" || params[attrTopicStr] != "retail/orders" {
		t.Fatalf("params = %v", params)
	}
	if _, ok := params[attrTopstr]; ok {
		t.Fatal("topstr should be mapped to topicStr for mqweb")
	}
}

func TestMapTopicRESTParameters_PubSubUppercase(t *testing.T) {
	t.Parallel()
	params := map[string]any{"pub": "enabled", "sub": "disabled"}
	mapTopicRESTParameters(params)
	if params["pub"] != "ENABLED" || params["sub"] != "DISABLED" {
		t.Fatalf("params = %v", params)
	}
}

func TestNormalizeTopicAttributes(t *testing.T) {
	t.Parallel()
	attrs := map[string]string{strings.ToLower(attrTopicStr): "retail/orders"}
	normalizeTopicAttributes(attrs)
	if attrs[attrTopstr] != "retail/orders" {
		t.Fatalf("attrs = %v", attrs)
	}
}

func TestNormalizeQueueAttributes(t *testing.T) {
	t.Parallel()
	t.Run("alias maps target to targq", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]string{"target": "APP.BASE"}
		normalizeQueueAttributes(attrs, mqadmin.QueueTypeAlias)
		if attrs["targq"] != "APP.BASE" {
			t.Fatalf("attrs = %v", attrs)
		}
	})
	t.Run("remote maps mqweb names", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]string{
			"remotequeue":       "REMOTE.Q",
			"remotemanager":     "QM2",
			"transmissionqueue": "XMIT.Q",
		}
		normalizeQueueAttributes(attrs, mqadmin.QueueTypeRemote)
		if attrs["rname"] != "REMOTE.Q" || attrs["rqmname"] != "QM2" || attrs["xmitq"] != "XMIT.Q" {
			t.Fatalf("attrs = %v", attrs)
		}
	})
	t.Run("local is no-op", func(t *testing.T) {
		t.Parallel()
		attrs := map[string]string{"maxdepth": "5000"}
		normalizeQueueAttributes(attrs, mqadmin.QueueTypeLocal)
		if attrs["maxdepth"] != "5000" {
			t.Fatalf("attrs = %v", attrs)
		}
	})
}

func TestQueueDisplayParametersByType(t *testing.T) {
	t.Parallel()
	if got := queueDisplayParameters(mqadmin.QueueTypeAlias); len(got) == 0 || got[0] != "targq" {
		t.Fatalf("alias display = %v", got)
	}
	if got := queueDisplayParameters(mqadmin.QueueTypeRemote); len(got) == 0 || got[0] != "rname" {
		t.Fatalf("remote display = %v", got)
	}
}

func TestDefineObjectParameters_InvalidNumericStaysString(t *testing.T) {
	t.Parallel()
	params := defineObjectParameters(map[string]string{attrMaxDepth: "not-a-number"}, queueNumericParameters)
	if params[attrMaxDepth] != "not-a-number" {
		t.Fatalf("params = %v", params)
	}
}

func TestQueueDisplayRequestUsesQualifier(t *testing.T) {
	t.Parallel()
	req := queueDisplayRequest(
		mqadmin.QueueSpec{Name: "APP.ALIAS", Type: mqadmin.QueueTypeAlias},
		queueDisplayParameters(mqadmin.QueueTypeAlias),
	)
	if req.Qualifier != "qalias" || req.Name != "APP.ALIAS" {
		t.Fatalf("request = %+v", req)
	}
}

func TestChannelDisplayRequestIncludesChltype(t *testing.T) {
	t.Parallel()
	req := channelDisplayRequest("ORDERS.APP", mqadmin.ChannelTypeSvrconn)
	if req.Qualifier != "channel" || req.Name != "ORDERS.APP" {
		t.Fatalf("request = %+v", req)
	}
	if req.Parameters["chltype"] != "svrconn" {
		t.Fatalf("chltype = %v", req.Parameters["chltype"])
	}
	if len(req.ResponseParameters) == 0 {
		t.Fatal("expected response parameters")
	}
}
