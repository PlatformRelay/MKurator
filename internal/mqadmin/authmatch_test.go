package mqadmin

import "testing"

func TestChannelAuthNeedsUpdate(t *testing.T) {
	t.Parallel()
	if !ChannelAuthNeedsUpdate(ChannelAuthSpec{}, nil) {
		t.Fatal("nil observed should need update")
	}
	desired := ChannelAuthSpec{
		ChannelName: "CH1",
		RuleType:    ChannelAuthRuleTypeAddressMap,
		Address:     "*",
		UserSource:  "CHANNEL",
		CheckClient: "REQUIRED",
		Description: "test",
	}
	observed := &ChannelAuthState{
		ChannelName: "CH1",
		RuleType:    ChannelAuthRuleTypeAddressMap,
		Address:     "*",
		UserSource:  "channel",
		CheckClient: "required",
		Description: "test",
	}
	if ChannelAuthNeedsUpdate(desired, observed) {
		t.Fatal("expected no update when attributes match (case-insensitive enums)")
	}
	observed.CheckClient = "ASQMGR"
	if !ChannelAuthNeedsUpdate(desired, observed) {
		t.Fatal("expected update when check client drifts")
	}
}

func TestAuthorityNeedsUpdate(t *testing.T) {
	t.Parallel()
	if !AuthorityNeedsUpdate(AuthoritySpec{Authorities: []string{"GET"}}, nil) {
		t.Fatal("nil observed should need update")
	}
	desired := AuthoritySpec{
		Profile:     "APP.ORDERS",
		ObjectType:  AuthorityObjectTypeQueue,
		Principal:   "app",
		Authorities: []string{"GET", "PUT"},
	}
	observed := &AuthorityState{
		Profile:     "APP.ORDERS",
		ObjectType:  AuthorityObjectTypeQueue,
		Principal:   "app",
		Authorities: []string{"put", "get"},
	}
	if AuthorityNeedsUpdate(desired, observed) {
		t.Fatal("expected no update when authority sets match")
	}
	observed.Authorities = []string{"GET"}
	if !AuthorityNeedsUpdate(desired, observed) {
		t.Fatal("expected update when authorities drift")
	}
}
