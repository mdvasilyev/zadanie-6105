package bid

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *Service) Patch(db *sql.DB, ctx *gin.Context) {
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

	if userExists := checkUserExistence(db, ctx, username); !userExists {
		return
	}

	var authorId string

	getAuthorIDQuery := `SELECT id FROM employee WHERE username = $1`

	err := db.QueryRow(getAuthorIDQuery, username).Scan(&authorId)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return
	}

	var bidPatch BidPatch
	if err := ctx.ShouldBindJSON(&bidPatch); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid request body"})
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	bid := Bid{}

	queryGet := "SELECT * FROM bid WHERE id = $1"

	err = tx.QueryRowContext(ctx, queryGet, bidId).Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.AuthorType, &bid.AuthorId, &bid.Version, &bid.CreatedAt)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	if authorId != bid.AuthorId {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return
	}

	changes := make(map[string]interface{})
	bidDiffValues := make(map[string]interface{})

	if bidPatch.Name != "" {
		changes["name"] = bidPatch.Name
		bidDiffValues["name"] = bidPatch.Name
		bid.Name = bidPatch.Name
	}

	if bidPatch.Description != "" {
		changes["description"] = bidPatch.Description
		bidDiffValues["description"] = bidPatch.Description
		bid.Description = bidPatch.Description
	}

	changes["version"] = bid.Version + 1

	var updates []string
	var params []interface{}
	paramCounter := 1

	for column, value := range changes {
		updates = append(updates, fmt.Sprintf("%s = $%d", column, paramCounter))
		params = append(params, value)
		paramCounter++
	}

	queryUpdate := fmt.Sprintf("UPDATE bid SET %s WHERE id = $%d", strings.Join(updates, ", "), paramCounter)
	queryDiff := "INSERT INTO bid_diff (id, name, description, status, tender_id, author_type, author_id, version, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	params = append(params, bidId)

	_, err = tx.ExecContext(ctx, queryUpdate, params...)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	_, err = tx.ExecContext(ctx, queryDiff, bid.Id, bid.Name, bid.Description, bid.Status, bid.TenderId, bid.AuthorType, bid.AuthorId, bid.Version+1, bid.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
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
		Status:      bid.Status,
		TenderId:    bid.TenderId,
		AuthorType:  bid.AuthorType,
		AuthorId:    bid.AuthorId,
		Version:     bid.Version + 1,
		CreatedAt:   bid.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusOK, returningBid)
}
