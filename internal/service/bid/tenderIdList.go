package bid

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *Service) TenderIdList(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	tenderId := ctx.Param("tenderId")
	if tenderId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"reason": "Tender ID is required"})
		return
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "5"))
	if err != nil || limit < 0 || limit > 50 {
		ctx.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid limit value"})
		return
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 || offset > 1<<31-1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid offset value"})
		return
	}

	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"reason": "Username is required"})
		return
	}

	if userExists := checkUserExistence(db, ctx, username); !userExists {
		return
	}

	if tenderExists := checkTenderExistence(db, ctx, tenderId); !tenderExists {
		return
	}

	authorId, ok := getAuthorId(db, ctx, username)
	if !ok {
		return
	}

	query := "SELECT id, name, description, status, tender_id, author_type, author_id, version, created_at FROM bid WHERE tender_id = $1 AND author_id = $2 ORDER BY name LIMIT $3 OFFSET $4"

	rows, err := db.Query(query, tenderId, authorId, limit, offset)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	defer rows.Close()

	bids, ok := extractBids(ctx, rows)
	if !ok {
		return
	}

	ctx.IndentedJSON(http.StatusOK, bids)
}
