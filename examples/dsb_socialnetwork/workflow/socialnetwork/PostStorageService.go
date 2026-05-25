package socialnetwork

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// The PostStorageService interface
type PostStorageService interface {
	// Stores a new `post` in the relevant backends.
	StorePost(ctx context.Context, reqID int64, post Post) error
	// Returns the post with id `postID`
	// If no post with id `postID` exists in the database, an error is returned.
	ReadPost(ctx context.Context, reqID int64, postID int64) (Post, error)
	// Returns a list of posts that have ids in the array `postIDs`.
	// No error is thrown if no post is found.
	ReadPosts(ctx context.Context, reqID int64, postIDs []int64) ([]Post, error)
}

// Implementation of [PostStorageService]
type PostStorageServiceImpl struct {
	postStorageCache backend.Cache
	postStorageDB    backend.NoSQLDatabase
	CacheHits        int64
	NumReqs          int64
	CacheMiss        int64
}

// Creates a [PostStorageService] instance that manages the post backends.
func NewPostStorageServiceImpl(ctx context.Context, postStorageCache backend.Cache, postStorageDB backend.NoSQLDatabase) (PostStorageService, error) {
	p := &PostStorageServiceImpl{postStorageCache: postStorageCache, postStorageDB: postStorageDB}
	return p, nil
}

// Implements PostStorageService interface
func (p *PostStorageServiceImpl) StorePost(ctx context.Context, reqID int64, post Post) error {
	collection, err := p.postStorageDB.GetCollection(ctx, "post_db", "post")
	if err != nil {
		return err
	}
	return collection.InsertOne(ctx, post)
}

// Implements PostStorageService interface
func (p *PostStorageServiceImpl) ReadPost(ctx context.Context, reqID int64, postID int64) (Post, error) {
	var post Post
	exists, err := p.postStorageCache.Get(ctx, strconv.FormatInt(postID, 10), &post)
	if err != nil {
		return post, err
	}
	if !exists {
		// Post was not in Cache, check DB!
		collection, err := p.postStorageDB.GetCollection(ctx, "post_db", "post")
		if err != nil {
			return post, err
		}
		query := bson.D{{Key: "PostID", Value: postID}}
		result, err := collection.FindOne(ctx, query)
		if err != nil {
			return post, err
		}
		res, err := result.One(ctx, &post)
		if !res || err != nil {
			return post, errors.New("Post doesn't exist in MongoDB")
		}
	}
	return post, nil
}

// Implements PostStorageService interface
func (p *PostStorageServiceImpl) ReadPosts(ctx context.Context, reqID int64, postIDs []int64) ([]Post, error) {
	unique_post_ids := make(map[int64]bool)
	for _, pid := range postIDs {
		unique_post_ids[pid] = true
	}
	if len(unique_post_ids) != len(postIDs) {
		return []Post{}, errors.New("Post Ids are duplicated")
	}
	var keys []string
	for _, pid := range postIDs {
		keys = append(keys, strconv.FormatInt(pid, 10))
	}
	values := make([]Post, len(keys))
	var retvals []interface{}
	for idx := range values {
		retvals = append(retvals, &values[idx])
	}
	p.postStorageCache.Mget(ctx, keys, retvals)
	var retposts []Post
	for _, post := range values {
		if post.PostID != 0 {
			delete(unique_post_ids, post.PostID)
			retposts = append(retposts, post)
		}
	}
	p.NumReqs += 1
	if len(unique_post_ids) != 0 {
		p.CacheMiss += 1
		var new_posts []Post
		var unique_pids []int64
		for k := range unique_post_ids {
			unique_pids = append(unique_pids, k)
		}
		collection, err := p.postStorageDB.GetCollection(ctx, "post_db", "post")
		if err != nil {
			return []Post{}, err
		}
		query_d := bson.D{
			{Key: "PostID", Value: bson.D{
				{Key: "$in", Value: unique_pids},
			}},
		}
		vals, err := collection.FindMany(ctx, query_d)
		if err != nil {
			log.Println(err)
			return []Post{}, err
		}
		err = vals.All(ctx, &new_posts)
		if err != nil {
			return []Post{}, err
		}
		retposts = append(retposts, new_posts...)
		for _, new_post := range new_posts {
			p.postStorageCache.Put(ctx, strconv.FormatInt(new_post.PostID, 10), new_post)
		}
	}
	if len(retposts) != len(postIDs) {
		return []Post{}, errors.New("Return set incomplete")
	}
	return retposts, nil
}
