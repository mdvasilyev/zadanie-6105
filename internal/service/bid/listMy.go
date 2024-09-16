package bid

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *Service) ListMy(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

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

	var authorId string

	getAuthorIDQuery := `SELECT id FROM employee WHERE username = $1`

	err = db.QueryRow(getAuthorIDQuery, username).Scan(&authorId)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return
	}

	query := "SELECT id, name, description, status, tender_id, author_type, author_id, version, created_at FROM bid WHERE author_id = $1 AND author_type = $2 ORDER BY name LIMIT $3 OFFSET $4"

	rows, err := db.Query(query, authorId, BidAuthorUser, limit, offset)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	defer rows.Close()

	var bids []Bid
	for rows.Next() {
		var b Bid
		err := rows.Scan(&b.Id, &b.Name, &b.Description, &b.Status, &b.TenderId, &b.AuthorType, &b.AuthorId, &b.Version, &b.CreatedAt)
		if err != nil {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
			return
		}
		bids = append(bids, b)
	}

	ctx.IndentedJSON(http.StatusOK, bids)
}
