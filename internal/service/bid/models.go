package bid

import (
	"github.com/google/uuid"
	"time"
)

type BidStatus string
type BidAuthor string

const (
	BidStatusCreated   BidStatus = "Created"
	BidStatusPublished BidStatus = "Published"
	BidStatusCancelled BidStatus = "Cancelled"
)

const (
	BidAuthorUser         BidAuthor = "User"
	BidAuthorOrganization BidAuthor = "Organization"
)

type Bid struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Status      BidStatus `json:"status"`
	TenderId    uuid.UUID `json:"tenderId" binding:"required"`
	AuthorType  BidAuthor `json:"authorType" binding:"required"`
	AuthorId    string    `json:"authorId" binding:"required"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"createdAt"`
}

type BidPatch struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
