package tender

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Status(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	tenderId := ctx.Param("tenderId")
	if tenderId == "" || len(tenderId) > 100 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid tenderId"})
		return
	}

	username := ctx.Query("username")

	query := "SELECT status, creator_username FROM tender WHERE id = $1"

	var status string
	var creatorUsername string

	err := db.QueryRow(query, tenderId).Scan(&status, &creatorUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": "Internal Server Error"})
		return
	}
	if status == string(TenderStatusPublished) {
		ctx.IndentedJSON(http.StatusOK, status)
		return
	}

	var userExists bool

	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`
	err = db.QueryRow(checkUserQuery, username).Scan(&userExists)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if !userExists {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return
	}

	if username != creatorUsername {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, status)
}
