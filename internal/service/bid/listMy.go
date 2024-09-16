package bid

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) ListMy(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	limit, ok := getLimit(ctx)
	if !ok {
		return
	}

	offset, ok := getOffset(ctx)
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

	query := "SELECT id, name, description, status, tender_id, author_type, author_id, version, created_at FROM bid WHERE author_id = $1 AND author_type = $2 ORDER BY name LIMIT $3 OFFSET $4"

	rows, err := db.Query(query, authorId, BidAuthorUser, limit, offset)
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
