package tender

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

	query := "SELECT * FROM tender WHERE creator_username = $1 ORDER BY name LIMIT $2 OFFSET $3"

	rows, err := db.Query(query, username, limit, offset)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	defer rows.Close()

	var tenders []Tender
	for rows.Next() {
		var t Tender
		err = rows.Scan(&t.Id, &t.Name, &t.Description, &t.Status, &t.ServiceType, &t.Version, &t.OrganizationId, &t.CreatorUsername, &t.CreatedAt)
		if err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
			return
		}
		tenders = append(tenders, t)
	}

	ctx.IndentedJSON(http.StatusOK, tenders)
}
