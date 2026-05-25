package tests

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_socialnetwork/workflow/socialnetwork"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplecache"
	"github.com/stretchr/testify/assert"
)

var homeTimelineServiceRegistry = registry.NewServiceRegistry[socialnetwork.HomeTimelineService]("hometimeline_service")

func init() {
	homeTimelineServiceRegistry.Register("local", func(ctx context.Context) (socialnetwork.HomeTimelineService, error) {
		cache, err := simplecache.NewSimpleCache(ctx)
		if err != nil {
			return nil, err
		}
		postStorageService, err := postStorageServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		socialGraphService, err := socialGraphServiceRegistry.Get(ctx)
		if err != nil {
			return nil, err
		}
		return socialnetwork.NewHomeTimelineServiceImpl(ctx, cache, postStorageService, socialGraphService)
	})
}

func TestHomeTimelineServiceReadEmpty(t *testing.T) {
	ctx := context.Background()
	service, err := homeTimelineServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	ids, err := service.ReadHomeTimeline(ctx, 0, 999, 0, 0)
	assert.NoError(t, err)
	assert.Empty(t, ids)
}

func TestHomeTimelineServiceReadInvalidRange(t *testing.T) {
	ctx := context.Background()
	service, err := homeTimelineServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	ids, err := service.ReadHomeTimeline(ctx, 0, 999, 5, 3)
	assert.NoError(t, err)
	assert.Empty(t, ids)
}

func TestHomeTimelineServiceWriteHomeTimeline(t *testing.T) {
	ctx := context.Background()
	socialGraph, err := socialGraphServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// user 61 follows user 60 — populates "60:followers" cache so WriteHomeTimeline can resolve followers
	err = socialGraph.InsertUser(ctx, 0, 60)
	assert.NoError(t, err)
	err = socialGraph.InsertUser(ctx, 0, 61)
	assert.NoError(t, err)
	err = socialGraph.Follow(ctx, 0, 61, 60)
	assert.NoError(t, err)

	service, err := homeTimelineServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	ts := time.Now().UnixNano()
	err = service.WriteHomeTimeline(ctx, 0, 5001, 60, ts, []int64{})
	assert.NoError(t, err)
}

func TestHomeTimelineServiceWriteAndRead(t *testing.T) {
	ctx := context.Background()
	socialGraph, err := socialGraphServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// user 71 follows user 70
	err = socialGraph.InsertUser(ctx, 0, 70)
	assert.NoError(t, err)
	err = socialGraph.InsertUser(ctx, 0, 71)
	assert.NoError(t, err)
	err = socialGraph.Follow(ctx, 0, 71, 70)
	assert.NoError(t, err)

	postStorage, err := postStorageServiceRegistry.Get(ctx)
	assert.NoError(t, err)
	err = postStorage.StorePost(ctx, 0, makePost(6001, 70, "user70", "hello"))
	assert.NoError(t, err)

	service, err := homeTimelineServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	ts := time.Now().UnixNano()
	err = service.WriteHomeTimeline(ctx, 0, 6001, 70, ts, []int64{})
	assert.NoError(t, err)

	// Post was written to user 71's home timeline
	ids, err := service.ReadHomeTimeline(ctx, 0, 71, 0, 10)
	assert.NoError(t, err)
	assert.Contains(t, ids, int64(6001))
}

func TestHomeTimelineServiceWriteWithMentionedUsers(t *testing.T) {
	ctx := context.Background()
	socialGraph, err := socialGraphServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	err = socialGraph.InsertUser(ctx, 0, 80)
	assert.NoError(t, err)
	err = socialGraph.InsertUser(ctx, 0, 81)
	assert.NoError(t, err)
	// user 81 follows user 80 (follower)
	err = socialGraph.Follow(ctx, 0, 81, 80)
	assert.NoError(t, err)

	postStorage, err := postStorageServiceRegistry.Get(ctx)
	assert.NoError(t, err)
	err = postStorage.StorePost(ctx, 0, makePost(7001, 80, "user80", "hi @82"))
	assert.NoError(t, err)

	service, err := homeTimelineServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	ts := time.Now().UnixNano()
	// user 82 is mentioned — also receives the post in their home timeline
	err = service.WriteHomeTimeline(ctx, 0, 7001, 80, ts, []int64{82})
	assert.NoError(t, err)

	// Both follower 81 and mentioned user 82 have the post
	ids81, err := service.ReadHomeTimeline(ctx, 0, 81, 0, 10)
	assert.NoError(t, err)
	assert.Contains(t, ids81, int64(7001))

	ids82, err := service.ReadHomeTimeline(ctx, 0, 82, 0, 10)
	assert.NoError(t, err)
	assert.Contains(t, ids82, int64(7001))
}
