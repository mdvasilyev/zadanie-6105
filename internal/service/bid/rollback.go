package bid

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Rollback(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	bidId, ok := getBidId(ctx)
	if !ok {
		return
	}

	newVersion, ok := getVersion(ctx)
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

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	currentVersion, ok := checkVersionAndUsername(tx, ctx, newVersion, authorId, bidId)
	if !ok {
		return
	}

	newBid, ok := getBidByIdAndVersion(tx, ctx, bidId, newVersion)
	if !ok {
		return
	}

	queryUpdate := "UPDATE bid SET name = $1, description = $2, status = $3, tender_id = $4, author_type = $5, author_id = $6, version = $7, created_at = $8 WHERE id = $9"

	_, err = tx.ExecContext(ctx, queryUpdate, newBid.Name, newBid.Description, newBid.Status, newBid.AuthorType, newBid.AuthorId, currentVersion+1, newBid.CreatedAt, newBid.Id)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Failed to rollback: %v", rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if !insertBidDiff(tx, ctx, newBid) {
		return
	}

	if err = tx.Commit(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	returningBid := Bid{
		Id:          newBid.Id,
		Name:        newBid.Name,
		Description: newBid.Description,
		Status:      newBid.Status,
		AuthorType:  newBid.AuthorType,
		AuthorId:    newBid.AuthorId,
		Version:     currentVersion + 1,
		CreatedAt:   newBid.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusOK, returningBid)
}
