package bid

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *Service) Rollback(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	bidId := ctx.Param("bidId")
	if bidId == "" || len(bidId) > 100 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid bidId"})
	}

	newVersion, err := strconv.Atoi(ctx.Param("version"))
	if err != nil || newVersion < 1 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Version must be >= 1"})
		return
	}

	username := ctx.Query("username")
	if username == "" {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	var userExists bool

	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`
	err = db.QueryRow(checkUserQuery, username).Scan(&userExists)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if !userExists {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return
	}

	var authorId string

	getAuthorIDQuery := `SELECT id FROM employee WHERE username = $1`

	err = db.QueryRow(getAuthorIDQuery, username).Scan(&authorId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	queryGet := "SELECT version, author_id FROM bid WHERE id = $1"

	var currentVersion int
	var creatorId string

	err = tx.QueryRowContext(ctx, queryGet, bidId).Scan(&currentVersion, &creatorId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Post failed: %v, unable to rollback: %v\n", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if newVersion >= currentVersion {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "No such a version. Latest version is " + strconv.Itoa(currentVersion)})
		return
	}

	if authorId != creatorId {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return
	}

	var newBid Bid

	queryGetDiff := "SELECT * FROM bid_diff WHERE id = $1 AND version = $2"

	err = tx.QueryRowContext(ctx, queryGetDiff, bidId, newVersion).Scan(&newBid.Id, &newBid.Name, &newBid.Description, &newBid.Status, &newBid.TenderId, &newBid.AuthorType, &newBid.AuthorId, &newBid.Version, &newBid.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Version not found"})
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
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

	queryInsertDiff := "INSERT INTO bid_diff (id, name, description, status, tender_id, author_type, author_id, version, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	_, err = tx.ExecContext(ctx, queryInsertDiff, newBid.Id, newBid.Name, newBid.Description, newBid.Status, newBid.AuthorType, newBid.AuthorId, newBid.Version+1, newBid.CreatedAt)
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
