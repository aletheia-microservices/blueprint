package tests

import (
	"context"
	"testing"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/dsb_socialnetwork/workflow/socialnetwork"
	"github.com/stretchr/testify/assert"
)

// TestWorkflow tests the full sequence of HTTP requests shown in the README

// Due to usage of simplenosqldb:
// - followWithUsername / UnfollowWithUsername are replaced with Follow / Unfollow by ID
// - the test does not cover composing post with images (due to $slice, $each, $position, etc.)
func TestWorkflow(t *testing.T) {
	ctx := context.Background()

	userService, err := userServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	socialGraph, err := socialGraphServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	composePost, err := composePostServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	homeTimeline, err := homeTimelineServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	userTimeline, err := userTimelineServiceRegistry.Get(ctx)
	assert.NoError(t, err)

	// Register users
	err = userService.RegisterUserWithId(ctx, 0, "John", "Doe", "alicedoe_rm", "secret123", 5001)
	assert.NoError(t, err)

	err = userService.RegisterUserWithId(ctx, 0, "Jane", "Smith", "bobsmith_rm", "secret456", 5002)
	assert.NoError(t, err)
	// Follow alicedoe (5001) -> bobsmith (5002)
	err = socialGraph.Follow(ctx, 0, 5001, 5002)
	assert.NoError(t, err)

	// Unfollow
	err = socialGraph.Unfollow(ctx, 0, 5001, 5002)
	assert.NoError(t, err)

	// Follow again
	err = socialGraph.Follow(ctx, 0, 5001, 5002)
	assert.NoError(t, err)

	// Sleep to avoid id collisions in unique id service
	time.Sleep(25 * time.Millisecond)

	// Compose a post (text only)
	postID1, mentionIDs, err := composePost.ComposePost(ctx, 0, "alicedoe_rm", 5001, "Hello from the social network", []int64{}, []string{}, socialnetwork.POST)
	assert.NoError(t, err)
	assert.Greater(t, postID1, int64(0))
	assert.Empty(t, mentionIDs)

	// Read home timeline of alicedoe (bobsmith hasn't posted, so empty)
	homePostIDs, err := homeTimeline.ReadHomeTimeline(ctx, 0, 5001, 0, 10)
	assert.NoError(t, err)
	assert.Empty(t, homePostIDs)

	// Read user timeline of alicedoe
	userPostIDs, err := userTimeline.ReadUserTimeline(ctx, 0, 5001, 0, 2)
	assert.NoError(t, err)
	assert.Len(t, userPostIDs, 1)
	assert.Contains(t, userPostIDs, postID1)
}
