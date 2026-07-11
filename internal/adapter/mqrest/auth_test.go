package mqrest

import (
	"errors"
	"strings"
	"testing"

	"github.com/platformrelay/mkurator/internal/mqadmin"
)

func TestBuildSetChannelAuthMQSC(t *testing.T) {
	cases := []struct {
		name   string
		spec   mqadmin.ChannelAuthSpec
		action string
		want   string
	}{
		{
			name: "addressmap replace",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "DEV.APP.SVRCONN.0TLS",
				RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
				Address:     "*",
				UserSource:  "CHANNEL",
				CheckClient: "REQUIRED",
				Description: "Allows connection via APP channel",
			},
			action: mqscActionReplace,
			want: "SET CHLAUTH('DEV.APP.SVRCONN.0TLS') TYPE(ADDRESSMAP) ADDRESS('*') " +
				"USERSRC(CHANNEL) CHCKCLNT(REQUIRED) DESCR('Allows connection via APP channel') ACTION(REPLACE)",
		},
		{
			name: "addressmap remove",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "DEV.APP.SVRCONN.0TLS",
				RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
				Address:     "*",
				UserSource:  "CHANNEL",
				CheckClient: "REQUIRED",
				Description: "ignored on remove",
			},
			action: mqscActionRemove,
			want:   "SET CHLAUTH('DEV.APP.SVRCONN.0TLS') TYPE(ADDRESSMAP) ADDRESS('*') ACTION(REMOVE)",
		},
		{
			name: "blockaddr replace",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "*",
				RuleType:    mqadmin.ChannelAuthRuleTypeBlockAddr,
				Address:     "192.0.2.1",
				Description: "block TEST-NET-1",
			},
			action: mqscActionReplace,
			want: "SET CHLAUTH('*') TYPE(BLOCKADDR) ADDRLIST('192.0.2.1') " +
				"DESCR('block TEST-NET-1') ACTION(REPLACE)",
		},
		{
			name: "blockaddr remove",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "*",
				RuleType:    mqadmin.ChannelAuthRuleTypeBlockAddr,
				Address:     "192.0.2.1",
				Description: "ignored on remove",
			},
			action: mqscActionRemove,
			want:   "SET CHLAUTH('*') TYPE(BLOCKADDR) ADDRLIST('192.0.2.1') ACTION(REMOVE)",
		},
		{
			name: "blockuser replace",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "ORDERS.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeBlockUser,
				UserList:    "nobody",
				Description: "deny nobody",
			},
			action: mqscActionReplace,
			want: "SET CHLAUTH('ORDERS.APP') TYPE(BLOCKUSER) USERLIST('nobody') " +
				"DESCR('deny nobody') ACTION(REPLACE)",
		},
		{
			name: "usermap replace",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "ORDERS.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeUserMap,
				ClientUser:  "johndoe",
				UserSource:  "MAP",
				McaUser:     "orders-app",
				Description: "map johndoe to orders-app",
			},
			action: mqscActionReplace,
			want: "SET CHLAUTH('ORDERS.APP') TYPE(USERMAP) CLNTUSER('johndoe') " +
				"USERSRC(MAP) MCAUSER('orders-app') DESCR('map johndoe to orders-app') ACTION(REPLACE)",
		},
		{
			name: "usermap remove",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "ORDERS.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeUserMap,
				ClientUser:  "johndoe",
				UserSource:  "MAP",
				McaUser:     "orders-app",
				Description: "ignored on remove",
			},
			action: mqscActionRemove,
			want:   "SET CHLAUTH('ORDERS.APP') TYPE(USERMAP) CLNTUSER('johndoe') ACTION(REMOVE)",
		},
		{
			name: "sslpeermap replace",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "ORDERS.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeSSLPeerMap,
				SSLPeerName: "CN=AppClient,O=MyOrg,C=US",
				UserSource:  "MAP",
				McaUser:     "orders-app",
				Description: "map cert DN to orders-app",
			},
			action: mqscActionReplace,
			want: "SET CHLAUTH('ORDERS.APP') TYPE(SSLPEERMAP) SSLPEER('CN=AppClient,O=MyOrg,C=US') " +
				"USERSRC(MAP) MCAUSER('orders-app') DESCR('map cert DN to orders-app') ACTION(REPLACE)",
		},
		{
			name: "sslpeermap remove",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "ORDERS.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeSSLPeerMap,
				SSLPeerName: "CN=AppClient,O=MyOrg,C=US",
				UserSource:  "MAP",
				McaUser:     "orders-app",
				Description: "ignored on remove",
			},
			action: mqscActionRemove,
			want:   "SET CHLAUTH('ORDERS.APP') TYPE(SSLPEERMAP) SSLPEER('CN=AppClient,O=MyOrg,C=US') ACTION(REMOVE)",
		},
		{
			name: "qmgrmap replace",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName:        "ORDERS.APP",
				RuleType:           mqadmin.ChannelAuthRuleTypeQMGRMap,
				RemoteQueueManager: "QM_PARTNER",
				UserSource:         "MAP",
				McaUser:            "orders-app",
				Description:        "map partner QM to orders-app",
			},
			action: mqscActionReplace,
			want: "SET CHLAUTH('ORDERS.APP') TYPE(QMGRMAP) QMNAME('QM_PARTNER') " +
				"USERSRC(MAP) MCAUSER('orders-app') DESCR('map partner QM to orders-app') ACTION(REPLACE)",
		},
		{
			name: "qmgrmap remove",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName:        "ORDERS.APP",
				RuleType:           mqadmin.ChannelAuthRuleTypeQMGRMap,
				RemoteQueueManager: "QM_PARTNER",
				UserSource:         "MAP",
				McaUser:            "orders-app",
				Description:        "ignored on remove",
			},
			action: mqscActionRemove,
			want:   "SET CHLAUTH('ORDERS.APP') TYPE(QMGRMAP) QMNAME('QM_PARTNER') ACTION(REMOVE)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := buildSetChannelAuthMQSC(tc.spec, tc.action)
			if err != nil {
				t.Fatalf("buildSetChannelAuthMQSC: %v", err)
			}
			if cmd != tc.want {
				t.Fatalf("got %q, want %q", cmd, tc.want)
			}
		})
	}
}

func TestBuildSetAuthorityMQSC(t *testing.T) {
	cmd, err := buildSetAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:     "APP.ORDERS",
		ObjectType:  mqadmin.AuthorityObjectTypeQueue,
		Principal:   "app",
		Authorities: []string{"GET", "PUT"},
	}, false)
	if err != nil {
		t.Fatalf("buildSetAuthorityMQSC: %v", err)
	}
	want := "SET AUTHREC PROFILE('APP.ORDERS') OBJTYPE(QUEUE) PRINCIPAL('app') AUTHADD(GET,PUT)"
	if cmd != want {
		t.Fatalf("got %q, want %q", cmd, want)
	}
}

func TestBuildSetAuthorityMQSCRemove(t *testing.T) {
	cmd, err := buildSetAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:    "APP.ORDERS",
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Group:      "apps",
	}, true)
	if err != nil {
		t.Fatalf("buildSetAuthorityMQSC: %v", err)
	}
	want := "SET AUTHREC PROFILE('APP.ORDERS') OBJTYPE(QUEUE) GROUP('apps') AUTHRMV(ALL)"
	if cmd != want {
		t.Fatalf("got %q, want %q", cmd, want)
	}
}

func TestBuildSetAuthorityMQSCObjectTypes(t *testing.T) {
	cases := []struct {
		name       string
		objectType mqadmin.AuthorityObjectType
		wantObj    string
	}{
		{"CHANNEL", mqadmin.AuthorityObjectTypeChannel, "CHANNEL"},
		{"TOPIC", mqadmin.AuthorityObjectTypeTopic, "TOPIC"},
		{"QMGR", mqadmin.AuthorityObjectTypeQMGR, "QMGR"},
		{"NAMESPAC", mqadmin.AuthorityObjectTypeNamespace, "NAMESPAC"},
		{"PROCESS", mqadmin.AuthorityObjectTypeProcess, "PROCESS"},
		{"NLIST", mqadmin.AuthorityObjectTypeNList, "NAMELIST"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := buildSetAuthorityMQSC(mqadmin.AuthoritySpec{
				Profile:     "APP.PROFILE",
				ObjectType:  tc.objectType,
				Principal:   "app",
				Authorities: []string{"CONNECT"},
			}, false)
			if err != nil {
				t.Fatalf("buildSetAuthorityMQSC: %v", err)
			}
			want := "SET AUTHREC PROFILE('APP.PROFILE') OBJTYPE(" + tc.wantObj +
				") PRINCIPAL('app') AUTHADD(CONNECT)"
			if cmd != want {
				t.Fatalf("got %q, want %q", cmd, want)
			}
		})
	}
}

func TestBuildSetChannelAuthMQSCValidation(t *testing.T) {
	_, err := buildSetChannelAuthMQSC(mqadmin.ChannelAuthSpec{}, mqscActionReplace)
	if err == nil {
		t.Fatal("expected error for empty channel name")
	}
	_, err = buildSetChannelAuthMQSC(mqadmin.ChannelAuthSpec{ChannelName: "CH1"}, mqscActionReplace)
	if err == nil {
		t.Fatal("expected error for empty rule type")
	}
}

func TestBuildSetAuthorityMQSCValidation(t *testing.T) {
	_, err := buildSetAuthorityMQSC(mqadmin.AuthoritySpec{}, false)
	if err == nil {
		t.Fatal("expected error for empty spec")
	}
	_, err = buildSetAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:    "APP.ORDERS",
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Principal:  "app",
	}, false)
	if err == nil {
		t.Fatal("expected error when authorities missing")
	}
	_, err = buildSetAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:     "APP.ORDERS",
		ObjectType:  mqadmin.AuthorityObjectTypeQueue,
		Principal:   "app",
		Group:       "apps",
		Authorities: []string{"GET"},
	}, false)
	if err == nil {
		t.Fatal("expected error when both principal and group set")
	}
}

func TestIsMQSCNotFound(t *testing.T) {
	if !isMQSCNotFound(errors.New("AMQ8147E: object not found")) {
		t.Fatal("expected not found")
	}
	if !isMQSCNotFound(errors.New("AMQ8958E: not found")) {
		t.Fatal("expected AMQ8958 not found")
	}
	if !isMQSCNotFound(mqadmin.ErrNotFound) {
		t.Fatal("expected ErrNotFound")
	}
	if isMQSCNotFound(errors.New("other error")) {
		t.Fatal("expected false")
	}
}

func TestMqscQuote(t *testing.T) {
	if got := mqscQuote("it's"); got != "it''s" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildDisplayChannelAuthMQSC(t *testing.T) {
	cases := []struct {
		name string
		spec mqadmin.ChannelAuthSpec
		want string
	}{
		{
			name: "addressmap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "DEV.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			},
			want: "DISPLAY CHLAUTH('DEV.APP') TYPE(ADDRESSMAP)",
		},
		{
			name: "addressmap with address",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "DEV.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
				Address:     "*",
			},
			want: "DISPLAY CHLAUTH('DEV.APP') TYPE(ADDRESSMAP) ADDRESS('*')",
		},
		{
			name: "sslpeermap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "ORDERS.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeSSLPeerMap,
				SSLPeerName: "CN=AppClient,O=MyOrg,C=US",
			},
			want: "DISPLAY CHLAUTH('ORDERS.APP') TYPE(SSLPEERMAP) SSLPEER('CN=AppClient,O=MyOrg,C=US')",
		},
		{
			name: "usermap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "ORDERS.APP",
				RuleType:    mqadmin.ChannelAuthRuleTypeUserMap,
				ClientUser:  "johndoe",
			},
			want: "DISPLAY CHLAUTH('ORDERS.APP') TYPE(USERMAP) CLNTUSER('johndoe')",
		},
		{
			name: "qmgrmap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName:        "ORDERS.APP",
				RuleType:           mqadmin.ChannelAuthRuleTypeQMGRMap,
				RemoteQueueManager: "QM_PARTNER",
			},
			want: "DISPLAY CHLAUTH('ORDERS.APP') TYPE(QMGRMAP) QMNAME('QM_PARTNER')",
		},
		{
			name: "blockaddr",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "*",
				RuleType:    mqadmin.ChannelAuthRuleTypeBlockAddr,
				Address:     "192.0.2.1",
			},
			want: "DISPLAY CHLAUTH('*') TYPE(BLOCKADDR) ADDRLIST",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := buildDisplayChannelAuthMQSC(tc.spec)
			if err != nil {
				t.Fatal(err)
			}
			if cmd != tc.want {
				t.Fatalf("got %q, want %q", cmd, tc.want)
			}
		})
	}
}

func TestBuildDisplayChannelAuthMQSCValidation(t *testing.T) {
	_, err := buildDisplayChannelAuthMQSC(mqadmin.ChannelAuthSpec{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildDisplayAuthorityMQSC(t *testing.T) {
	cmd, err := buildDisplayAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:    "APP.ORDERS",
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Principal:  "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "DISPLAY AUTHREC PROFILE('APP.ORDERS') OBJTYPE(QUEUE) PRINCIPAL('app')"
	if cmd != want {
		t.Fatalf("got %q, want %q", cmd, want)
	}
}

func TestAuthorityStateFromAttributes(t *testing.T) {
	spec := mqadmin.AuthoritySpec{
		Profile:    "APP.ORDERS",
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Principal:  "app",
	}
	state := authorityStateFromAttributes(spec, map[string]string{"authlist": "GET, PUT"})
	if len(state.Authorities) != 2 || state.Authorities[0] != "GET" || state.Authorities[1] != "PUT" {
		t.Fatalf("authorities = %v", state.Authorities)
	}
}

func TestBuildDisplayAuthorityMQSCTopicObject(t *testing.T) {
	cmd, err := buildDisplayAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:    "SYSTEM.ADMIN.TOPIC",
		ObjectType: mqadmin.AuthorityObjectTypeTopic,
		Principal:  "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cmd, "OBJTYPE(TOPIC)") {
		t.Fatalf("cmd = %q", cmd)
	}
}

func TestBuildDisplayAuthorityMQSCChannelObject(t *testing.T) {
	cmd, err := buildDisplayAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:    "ORDERS.APP",
		ObjectType: mqadmin.AuthorityObjectTypeChannel,
		Principal:  "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "DISPLAY AUTHREC PROFILE('ORDERS.APP') OBJTYPE(CHANNEL) PRINCIPAL('app')"
	if cmd != want {
		t.Fatalf("got %q, want %q", cmd, want)
	}
}

func TestBuildDisplayAuthorityMQSCNListObject(t *testing.T) {
	cmd, err := buildDisplayAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:    "APP.NLIST",
		ObjectType: mqadmin.AuthorityObjectTypeNList,
		Principal:  "app",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "DISPLAY AUTHREC PROFILE('APP.NLIST') OBJTYPE(NAMELIST) PRINCIPAL('app')"
	if cmd != want {
		t.Fatalf("got %q, want %q", cmd, want)
	}
}

func TestBuildDisplayAuthorityMQSCGroup(t *testing.T) {
	cmd, err := buildDisplayAuthorityMQSC(mqadmin.AuthoritySpec{
		Profile:    "APP.ORDERS",
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Group:      "apps",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cmd, "GROUP('apps')") {
		t.Fatalf("cmd = %q", cmd)
	}
}

func TestBuildDisplayAuthorityMQSCValidation(t *testing.T) {
	_, err := buildDisplayAuthorityMQSC(mqadmin.AuthoritySpec{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChannelAuthStateFromAttributes(t *testing.T) {
	cases := []struct {
		name   string
		spec   mqadmin.ChannelAuthSpec
		attrs  map[string]string
		assert func(t *testing.T, state *mqadmin.ChannelAuthState)
	}{
		{
			name: "addressmap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "CH1",
				RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
			},
			attrs: map[string]string{
				"address": "*", "usersrc": "CHANNEL", "chckclnt": "REQUIRED", "descr": "d",
			},
			assert: func(t *testing.T, state *mqadmin.ChannelAuthState) {
				t.Helper()
				if state.Description != "d" {
					t.Fatalf("state = %+v", state)
				}
			},
		},
		{
			name: "sslpeermap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "CH1",
				RuleType:    mqadmin.ChannelAuthRuleTypeSSLPeerMap,
			},
			attrs: map[string]string{
				attrSslPeer: "CN=AppClient,O=MyOrg,C=US", attrMcaUser: "orders-app", "usersrc": "MAP", "descr": "map",
			},
			assert: func(t *testing.T, state *mqadmin.ChannelAuthState) {
				t.Helper()
				if state.SSLPeerName != "CN=AppClient,O=MyOrg,C=US" || state.McaUser != "orders-app" ||
					state.UserSource != "MAP" {
					t.Fatalf("state = %+v", state)
				}
			},
		},
		{
			name: "usermap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "CH1",
				RuleType:    mqadmin.ChannelAuthRuleTypeUserMap,
			},
			attrs: map[string]string{
				"clntuser": "johndoe", attrMcaUser: "orders-app", "usersrc": "MAP", "descr": "map",
			},
			assert: func(t *testing.T, state *mqadmin.ChannelAuthState) {
				t.Helper()
				if state.ClientUser != "johndoe" || state.McaUser != "orders-app" || state.UserSource != "MAP" {
					t.Fatalf("state = %+v", state)
				}
			},
		},
		{
			name: "qmgrmap",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "CH1",
				RuleType:    mqadmin.ChannelAuthRuleTypeQMGRMap,
			},
			attrs: map[string]string{
				attrQmName: "QM_PARTNER", attrMcaUser: "orders-app", "usersrc": "MAP", "descr": "map",
			},
			assert: func(t *testing.T, state *mqadmin.ChannelAuthState) {
				t.Helper()
				if state.RemoteQueueManager != "QM_PARTNER" || state.McaUser != "orders-app" ||
					state.UserSource != "MAP" {
					t.Fatalf("state = %+v", state)
				}
			},
		},
		{
			name: "blockaddr",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "*",
				RuleType:    mqadmin.ChannelAuthRuleTypeBlockAddr,
			},
			attrs: map[string]string{
				"addrlist": "192.0.2.1", "descr": "block",
			},
			assert: func(t *testing.T, state *mqadmin.ChannelAuthState) {
				t.Helper()
				if state.Address != "192.0.2.1" || state.Description != "block" {
					t.Fatalf("state = %+v", state)
				}
			},
		},
		{
			name: "blockuser",
			spec: mqadmin.ChannelAuthSpec{
				ChannelName: "CH1",
				RuleType:    mqadmin.ChannelAuthRuleTypeBlockUser,
			},
			attrs: map[string]string{
				"userlist": "nobody", "descr": "block",
			},
			assert: func(t *testing.T, state *mqadmin.ChannelAuthState) {
				t.Helper()
				if state.UserList != "nobody" || state.Description != "block" {
					t.Fatalf("state = %+v", state)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := channelAuthStateFromAttributes(tc.spec, tc.attrs)
			tc.assert(t, state)
		})
	}
}
