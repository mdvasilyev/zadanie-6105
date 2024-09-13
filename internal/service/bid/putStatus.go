package bid

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) PutStatus(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	bidId := ctx.Param("bidId")
	if bidId == "" || len(bidId) > 100 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid bidId"})
		return
	}

	username := ctx.Query("username")
	if username == "" {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	var userExists bool

	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`
	err := db.QueryRow(checkUserQuery, username).Scan(&userExists)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if !userExists {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return
	}

	var authorId string

	queryAuthorId := "SELECT id FROM employee WHERE username = $1"

	err = db.QueryRow(queryAuthorId, username).Scan(&authorId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	newStatus := ctx.Query("status")
	if err := validateStatus(newStatus); newStatus == "" || err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid status"})
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	queryGet := "SELECT * FROM bid WHERE id = $1"

	var bid Bid

	err = tx.QueryRowContext(ctx, queryGet, bidId).Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.AuthorType, &bid.AuthorId, &bid.Version, &bid.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
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

	queryInsertDiff := "INSERT INTO bid_diff (id, name, description, status, tender_id, author_type, author_id, version, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	_, err = tx.ExecContext(ctx, queryInsertDiff, bid.Id, bid.Name, bid.Description, newStatus, bid.TenderId, bid.AuthorType, bid.AuthorId, bid.Version+1, bid.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Failed to rollback: %v", rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
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
