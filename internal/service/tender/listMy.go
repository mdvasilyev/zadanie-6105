package tender

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

	query := "SELECT * FROM tender WHERE creator_username = $1 ORDER BY name LIMIT $2 OFFSET $3"

	rows, err := db.Query(query, username, limit, offset)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	defer rows.Close()

	tenders, ok := extractTenders(ctx, rows)
	if !ok {
		return
	}

	ctx.IndentedJSON(http.StatusOK, tenders)
}
