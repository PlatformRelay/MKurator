//go:build integration

package mq

import (
	"context"
	"errors"
	"testing"

	"github.com/konih/kurator/internal/mqadmin"
)

func TestIntegration_GetChannelAuth(t *testing.T) {
	requireIntegration(t)
	ctx := testContext(t)
	channel := channelNameForTest(t.Name())

	c, err := newIntegrationClient()
	if err != nil {
		t.Fatal(err)
	}

	chSpec := mqadmin.ChannelSpec{
		Name: channel,
		Type: mqadmin.ChannelTypeSvrconn,
		Attributes: map[string]string{
			"trptype": "tcp",
		},
	}
	authSpec := mqadmin.ChannelAuthSpec{
		ChannelName: channel,
		RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
		Address:     "*",
		UserSource:  "CHANNEL",
		CheckClient: "REQUIRED",
		Description: "integration get path",
	}
	t.Cleanup(func() {
		_ = c.DeleteChannelAuth(context.Background(), authSpec)
		_ = c.DeleteChannel(context.Background(), chSpec)
	})

	if err := c.DefineChannel(ctx, chSpec); err != nil {
		t.Fatalf("DefineChannel: %v", err)
	}

	if err := c.SetChannelAuth(ctx, authSpec); err != nil {
		t.Fatalf("SetChannelAuth: %v", err)
	}

	state, err := c.GetChannelAuth(ctx, authSpec)
	if err != nil {
		t.Fatalf("GetChannelAuth: %v", err)
	}
	if state.Address != "*" || state.UserSource != "CHANNEL" {
		t.Fatalf("state = %+v", state)
	}
}

func TestIntegration_GetChannelAuth_NotFound(t *testing.T) {
	requireIntegration(t)
	ctx := testContext(t)

	c, err := newIntegrationClient()
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.GetChannelAuth(ctx, mqadmin.ChannelAuthSpec{
		ChannelName: channelNameForTest(t.Name()+".missing"),
		RuleType:    mqadmin.ChannelAuthRuleTypeAddressMap,
	})
	if err == nil || !errors.Is(err, mqadmin.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestIntegration_GetAuthority(t *testing.T) {
	requireIntegration(t)
	ctx := testContext(t)
	profile := queueNameForTest(t.Name())

	c, err := newIntegrationClient()
	if err != nil {
		t.Fatal(err)
	}

	queueSpec := mqadmin.QueueSpec{
		Name:       profile,
		Type:       mqadmin.QueueTypeLocal,
		Attributes: map[string]string{"maxdepth": "100"},
	}
	authSpec := mqadmin.AuthoritySpec{
		Profile:     profile,
		ObjectType:  mqadmin.AuthorityObjectTypeQueue,
		Principal:   "app",
		Authorities: []string{"GET", "PUT"},
	}
	t.Cleanup(func() {
		_ = c.DeleteAuthority(context.Background(), authSpec)
		_ = c.DeleteQueue(context.Background(), queueSpec)
	})

	if err := c.DefineQueue(ctx, queueSpec); err != nil {
		t.Fatalf("DefineQueue: %v", err)
	}
	if err := c.SetAuthority(ctx, authSpec); err != nil {
		t.Fatalf("SetAuthority: %v", err)
	}

	state, err := c.GetAuthority(ctx, authSpec)
	if err != nil {
		t.Fatalf("GetAuthority: %v", err)
	}
	if len(state.Authorities) < 2 {
		t.Fatalf("authorities = %v", state.Authorities)
	}
}

func TestIntegration_GetAuthority_NotFound(t *testing.T) {
	requireIntegration(t)
	ctx := testContext(t)

	c, err := newIntegrationClient()
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.GetAuthority(ctx, mqadmin.AuthoritySpec{
		Profile:    queueNameForTest(t.Name() + ".missing"),
		ObjectType: mqadmin.AuthorityObjectTypeQueue,
		Principal:  "nobody",
	})
	if err == nil || !errors.Is(err, mqadmin.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
