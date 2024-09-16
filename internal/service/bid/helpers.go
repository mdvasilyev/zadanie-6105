package bid

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func validateStatus(status string) error {
	switch BidStatus(status) {
	case BidStatusCreated, BidStatusPublished, BidStatusCancelled:
		return nil
	}
	return errors.New("invalid status")
}

func checkUserExistence(db *sql.DB, ctx *gin.Context, username string) bool {
	var userExists bool

	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`
	err := db.QueryRow(checkUserQuery, username).Scan(&userExists)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return false
	}
	if !userExists {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return false
	}
	return true
}

func getAuthorId(db *sql.DB, ctx *gin.Context, username string) (string, bool) {
	var authorId string

	getAuthorIDQuery := `SELECT id FROM employee WHERE username = $1`

	err := db.QueryRow(getAuthorIDQuery, username).Scan(&authorId)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return "", false
	}
	return authorId, true
}

func extractBids(ctx *gin.Context, rows *sql.Rows) ([]Bid, bool) {
	var bids []Bid
	for rows.Next() {
		var b Bid
		err := rows.Scan(&b.Id, &b.Name, &b.Description, &b.Status, &b.TenderId, &b.AuthorType, &b.AuthorId, &b.Version, &b.CreatedAt)
		if err != nil {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bids not found"})
			return nil, false
		}
		bids = append(bids, b)
	}
	return bids, true
}
