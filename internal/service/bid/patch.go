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

	bid, ok := getBidById(tx, ctx, bidId)
	if !ok {
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
		Status:      bid.Status,
		TenderId:    bid.TenderId,
		AuthorType:  bid.AuthorType,
		AuthorId:    bid.AuthorId,
		Version:     bid.Version + 1,
		CreatedAt:   bid.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusOK, returningBid)
}
