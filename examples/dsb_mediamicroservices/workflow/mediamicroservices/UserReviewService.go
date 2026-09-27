package mediamicroservices

import (
	"context"
	"errors"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type UserReviewData struct {
	ReviewID  int64 `bson:"ReviewID"`
	Timestamp int64 `bson:"Timestamp"`
}

type UserReview struct {
	UserID  int64            `bson:"UserID"`
	Reviews []UserReviewData `bson:"Reviews"`
}

type UserReviewService interface {
	UploadUserReview(ctx context.Context, reqID int64, userID int64, reviewID int64, timestamp int64) error
	ReadUserReviews(ctx context.Context, reqID int64, userID int64, start int, stop int) ([]Review, error)
}

type UserReviewServiceImpl struct {
	database             backend.NoSQLDatabase
	cache                backend.Cache
	reviewStorageService ReviewStorageService
}

func NewUserReviewServiceImpl(ctx context.Context, database backend.NoSQLDatabase, cache backend.Cache, reviewStorageService ReviewStorageService) (UserReviewService, error) {
	s := &UserReviewServiceImpl{database: database, cache: cache, reviewStorageService: reviewStorageService}
	return s, nil
}

func (s *UserReviewServiceImpl) UploadUserReview(ctx context.Context, reqID int64, userID int64, reviewID int64, timestamp int64) error {
	collection, err := s.database.GetCollection(ctx, "user_review_db", "user_review")
	if err != nil {
		return err
	}

	var movieReview UserReview
	query := bson.D{{Key: "UserID", Value: userID}}
	result, err := collection.FindOne(ctx, query)
	if err != nil {
		return err
	}
	found, err := result.One(ctx, &movieReview)
	if err != nil {
		return err
	}

	if !found {
		movieReview = UserReview{
			UserID: userID,
			Reviews: []UserReviewData{{
				ReviewID:  reviewID,
				Timestamp: timestamp,
			}},
		}
		err := collection.InsertOne(ctx, movieReview)
		if err != nil {
			return err
		}
	} else {
		update := bson.D{
			{Key: "$push", Value: bson.D{
				{Key: "Reviews", Value: bson.D{
					{Key: "$each", Value: bson.A{
						bson.D{
							{Key: "ReviewID", Value: reviewID},
							{Key: "Timestamp", Value: timestamp},
						},
					}},
					{Key: "$position", Value: 0},
				}},
			}},
		}
		_, err = collection.UpdateMany(ctx, query, update)
		if err != nil {
			return errors.New("Failed to insert user timeline user to Database")
		}
	}

	var reviews []UserReviewData
	// Ignore error check for Get!
	userIDStr := strconv.FormatInt(userID, 10)
	_, err = s.cache.Get(ctx, userIDStr, &reviews)
	if err != nil {
		return err
	}
	reviews = append(reviews, UserReviewData{ReviewID: reviewID, Timestamp: timestamp})
	return s.cache.Put(ctx, userIDStr, reviews)
}

func (s *UserReviewServiceImpl) ReadUserReviews(ctx context.Context, reqID int64, userID int64, start int, stop int) ([]Review, error) {
	if stop <= start || start < 0 {
		return nil, nil
	}

	var reviews []Review
	userIDStr := strconv.FormatInt(userID, 10)
	exists, err := s.cache.Get(ctx, userIDStr, &reviews)
	if err != nil {
		return nil, err
	}

	var reviewIds []int64
	seen_reviews_ids := make(map[int64]bool)
	for _, review := range reviews {
		reviewIds = append(reviewIds, review.ReviewID)
		seen_reviews_ids[review.ReviewID] = true
	}
	db_start := start + len(reviewIds)
	var new_reviews_ids []int64
	if db_start < stop {
		collection, err := s.database.GetCollection(ctx, "user_review_db", "user_review")
		if err != nil {
			return nil, err
		}

		query := bson.D{{Key: "UserID", Value: userID}}

		projection := bson.D{
			{Key: "posts", Value: bson.D{
				{Key: "$slice", Value: bson.A{0, stop}},
			}},
		}

		cursor, err := collection.FindOne(ctx, query, projection)
		if err != nil {
			return nil, err
		}
		var userReview UserReview
		exists, err = cursor.One(ctx, &userReview)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("Failed to find posts in database")
		}
		for _, review := range userReview.Reviews {
			// Avoid duplicated reviews
			if _, ok := seen_reviews_ids[review.ReviewID]; ok {
				continue
			}
			new_reviews_ids = append(new_reviews_ids, review.ReviewID)
		}
	}

	reviewIds = append(new_reviews_ids, reviewIds...)

	// update reviews []Review
	reviews, err = s.reviewStorageService.ReadReviews(ctx, reqID, reviewIds)
	if err != nil {
		return nil, err
	}

	if len(new_reviews_ids) > 0 {
		err := s.cache.Put(ctx, userIDStr, reviews)
		if err != nil {
			return nil, err
		}
	}

	return reviews, nil
}
