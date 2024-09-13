package bid

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Status(db *sql.DB, ctx *gin.Context) {
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
	var authorIdInDb string

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

	var status string
	var tenderId string

	query := "SELECT status, tender_id, author_id FROM bid WHERE id = $1"

	err = db.QueryRow(query, bidId).Scan(&status, &tenderId, &authorIdInDb)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Bid not found"})
		return
	}

	if authorId == authorIdInDb {
		ctx.IndentedJSON(http.StatusOK, status)
		return
	}

	var responsibleExists bool

	queryResponsible := `
    SELECT EXISTS(
        SELECT 1
        FROM organization_responsible
        WHERE user_id = $1
        AND organization_id = (
            SELECT organization_id FROM tender WHERE id = $2
        )
    )`

	err = db.QueryRow(queryResponsible, authorId, tenderId).Scan(&responsibleExists)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": "Internal Server Error"})
		return
	}

	if responsibleExists {
		ctx.IndentedJSON(http.StatusOK, status)
		return
	}

	ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
}
