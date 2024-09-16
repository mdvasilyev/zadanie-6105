package bid

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
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

	query := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`

	err := db.QueryRow(query, username).Scan(&userExists)
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

func checkTenderExistence(db *sql.DB, ctx *gin.Context, tenderId string) bool {
	var tenderExists bool

	query := `SELECT EXISTS(SELECT 1 FROM tender WHERE id = $1)`

	err := db.QueryRow(query, tenderId).Scan(&tenderExists)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return false
	}

	if !tenderExists {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return false
	}

	return true
}

func getAuthorId(db *sql.DB, ctx *gin.Context, username string) (string, bool) {
	var authorId string

	query := `SELECT id FROM employee WHERE username = $1`

	err := db.QueryRow(query, username).Scan(&authorId)
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

func checkVersionAndUsername(tx *sql.Tx, ctx *gin.Context, version int, authorId string, bidId string) (int, bool) {
	query := "SELECT version, author_id FROM bid WHERE id = $1"

	var currentVersion int
	var creatorId string

	err := tx.QueryRowContext(ctx, query, bidId).Scan(&currentVersion, &creatorId)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return 0, false
	}

	if version >= currentVersion {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "No such a version. Latest version is " + strconv.Itoa(currentVersion)})
		return 0, false
	}

	if authorId != creatorId {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return 0, false
	}

	return currentVersion, true
}

func insertBidDiff(tx *sql.Tx, ctx *gin.Context, bid Bid) bool {
	query := "INSERT INTO bid_diff (id, name, description, status, tender_id, author_type, author_id, version, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	_, err := tx.ExecContext(ctx, query, bid.Id, bid.Name, bid.Description, bid.Status, bid.AuthorType, bid.AuthorId, bid.Version+1, bid.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Failed to rollback: %v", rollbackErr)})
			return false
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return false
	}

	return true
}
