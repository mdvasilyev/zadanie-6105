package bid

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Status(db *sql.DB, ctx *gin.Context) {
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

	var authorIdInDb string
	var status string
	var tenderId string

	query := "SELECT status, tender_id, author_id FROM bid WHERE id = $1"

	err := db.QueryRow(query, bidId).Scan(&status, &tenderId, &authorIdInDb)
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
