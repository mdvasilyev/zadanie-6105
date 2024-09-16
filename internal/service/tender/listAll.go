package tender

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *Service) ListAll(db *sql.DB, ctx *gin.Context) {
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

	serviceType := ctx.Query("service_type")

	query := "SELECT * FROM tender"
	paramCounter := 1

	if serviceType != "" {
		if err := validateServiceType(serviceType); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"reason": "Invalid service type value"})
			return
		}
		query += fmt.Sprintf("\nWHERE service_type = $%d AND status = $%d ORDER BY name LIMIT $%d OFFSET $%d", paramCounter, paramCounter+1, paramCounter+2, paramCounter+3)
		paramCounter += 2
	} else {
		query += fmt.Sprintf("\nWHERE status = $%d ORDER BY name ASC LIMIT $%d OFFSET $%d", paramCounter, paramCounter+1, paramCounter+2)
	}

	var rows *sql.Rows
	if serviceType == "" {
		rows, err = db.Query(query, TenderStatusPublished, limit, offset)
	} else {
		rows, err = db.Query(query, TenderServiceType(serviceType), TenderStatusPublished, limit, offset)
	}
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
