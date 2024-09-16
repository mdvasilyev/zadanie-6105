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

func insertBid(tx *sql.Tx, ctx *gin.Context, bid Bid) (Bid, bool) {
	query := "INSERT INTO bid (name, description, status, tender_id, author_type, author_id, version) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at"

	err := tx.QueryRowContext(ctx, query, bid.Name, bid.Description, BidStatusCreated, bid.TenderId, bid.AuthorType, bid.AuthorId, 1).Scan(&bid.Id, &bid.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Failed to rollback: %v", rollbackErr)})
			return bid, false
		}
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid tenderId or authorType or authorId"})
		return bid, false
	}

	return bid, true
}

func insertBidDiff(tx *sql.Tx, ctx *gin.Context, bid Bid) bool {
	query := "INSERT INTO bid_diff (id, name, description, status, tender_id, author_type, author_id, version, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	if bid.Status == "" {
		bid.Status = BidStatusCreated
	}

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

func getBidById(tx *sql.Tx, ctx *gin.Context, bidId string) (Bid, bool) {
	query := "SELECT * FROM bid WHERE id = $1"

	var bid Bid

	err := tx.QueryRowContext(ctx, query, bidId).Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.AuthorType, &bid.AuthorId, &bid.Version, &bid.CreatedAt)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return bid, false
	}

	return bid, true
}

func getBidByIdAndVersion(tx *sql.Tx, ctx *gin.Context, bidId string, version int) (Bid, bool) {
	var bid Bid

	query := "SELECT * FROM bid_diff WHERE id = $1 AND version = $2"

	err := tx.QueryRowContext(ctx, query, bidId, version).Scan(&bid.Id, &bid.Name, &bid.Description, &bid.Status, &bid.TenderId, &bid.AuthorType, &bid.AuthorId, &bid.Version, &bid.CreatedAt)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Version not found"})
		return bid, false
	}

	return bid, true
}

func getTenderId(ctx *gin.Context) (string, bool) {
	tenderId := ctx.Param("tenderId")

	if tenderId == "" || len(tenderId) > 100 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid tenderId"})
		return "", false
	}

	return tenderId, true
}

func getUsername(ctx *gin.Context) (string, bool) {
	username := ctx.Query("username")

	if username == "" {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return "", false
	}

	return username, true
}

func getLimit(ctx *gin.Context) (int, bool) {
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "5"))

	if err != nil || limit < 0 || limit > 50 {
		ctx.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid limit value"})
		return 0, false
	}

	return limit, true
}

func getOffset(ctx *gin.Context) (int, bool) {
	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	if err != nil || offset < 0 || offset > 1<<31-1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid offset value"})
		return 0, false
	}

	return offset, true
}

func getBidId(ctx *gin.Context) (string, bool) {
	bidId := ctx.Param("bidId")

	if bidId == "" || len(bidId) > 100 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid bidId"})
		return "", false
	}

	return bidId, true
}

func getStatus(ctx *gin.Context) (string, bool) {
	status := ctx.Query("status")

	if err := validateStatus(status); status == "" || err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid status"})
		return "", false
	}

	return status, true
}

func getVersion(ctx *gin.Context) (int, bool) {
	version, err := strconv.Atoi(ctx.Param("version"))

	if err != nil || version < 1 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Version must be >= 1"})
		return 0, false
	}

	return version, true
}
