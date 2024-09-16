package bid

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) PutStatus(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	bidId, ok := getBidId(ctx)
	if !ok {
		return
	}

	username, ok := getUsername(ctx)
	if !ok {
		return
	}

	if userExists := checkUserExistence(db, ctx, username); !userExists {
		return
	}

	authorId, ok := getAuthorId(db, ctx, username)
	if !ok {
		return
	}

	newStatus, ok := getStatus(ctx)
	if !ok {
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	bid, ok := getBidById(tx, ctx, bidId)
	if !ok {
		return
	}

	if newStatus == string(bid.Status) {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": fmt.Sprintf("Status is already %v", newStatus)})
		return
	}

	if authorId != bid.AuthorId {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return
	}

	queryUpdate := "UPDATE bid SET status = $1, version = $2 WHERE id = $3"

	_, err = tx.ExecContext(ctx, queryUpdate, newStatus, bid.Version+1, bidId)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if !insertBidDiff(tx, ctx, bid) {
		return
	}

	if err = tx.Commit(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	returningBid := Bid{
		Id:          bid.Id,
		Name:        bid.Name,
		Description: bid.Description,
		Status:      BidStatus(newStatus),
		TenderId:    bid.TenderId,
		AuthorType:  bid.AuthorType,
		AuthorId:    bid.AuthorId,
		Version:     bid.Version + 1,
		CreatedAt:   bid.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusOK, returningBid)
}
