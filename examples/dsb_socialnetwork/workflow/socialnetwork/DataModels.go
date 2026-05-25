package socialnetwork

// The format of a user stored in the user database
type User struct {
	UserID    int64  `bson:"UserID"`
	FirstName string `bson:"FirstName"`
	LastName  string `bson:"LastName"`
	Username  string `bson:"Username"`
	PwdHashed string `bson:"PwdHashed"`
	Salt      string `bson:"Salt"`
}

// The format of a media stored as part of a post.
type Media struct {
	MediaID   int64  `bson:"MediaID"`
	MediaType string `bson:"MediaType"`
}

// The format of a url stored in the url-shorten database
type URL struct {
	ShortenedUrl string `bson:"ShortenedUrl"`
	ExpandedUrl  string `bson:"ExpandedUrl"`
}

// The format of a usermention stored as part of a post
type UserMention struct {
	UserID   int64  `bson:"UserID"`
	Username string `bson:"Username"`
}

// The format of a creator stored as part of a post
type Creator struct {
	UserID   int64  `bson:"UserID"`
	Username string `bson:"Username"`
}

// The type of the post.
type PostType int64

// Enums aren't supported atm. So just use integers instead.
const (
	POST int64 = iota
//	REPOST
//	REPLY
//	DM
)

type Post struct {
	PostID       int64         `bson:"PostID"`
	Creator      Creator       `bson:"Creator"`
	ReqID        int64         `bson:"ReqID"`
	Text         string        `bson:"Text"`
	UserMentions []UserMention `bson:"UserMentions"`
	Medias       []Media       `bson:"Medias"`
	Urls         []URL         `bson:"Urls"`
	Timestamp    int64         `bson:"Timestamp"`
	PostType     int64         `bson:"PostType"`
}
