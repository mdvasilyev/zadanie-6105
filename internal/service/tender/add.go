package tender

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Add(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	tender := Tender{Version: 1}
	if err = ctx.ShouldBindJSON(&tender); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid request data"})
		return
	}

	queryTender := "INSERT INTO tender (name, description, status, service_type, version, organization_id, creator_username) VALUES ($1, $2, $3, $4, 1, $5, $6) RETURNING id, created_at"

	err = tx.QueryRowContext(ctx, queryTender, tender.Name, tender.Description, TenderStatusCreated, tender.ServiceType, tender.OrganizationId, tender.CreatorUsername).Scan(&tender.Id, &tender.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if !insertTenderDiff(tx, ctx, tender) {
		return
	}

	if err = tx.Commit(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	returningTender := Tender{
		Id:              tender.Id,
		Name:            tender.Name,
		Description:     tender.Description,
		ServiceType:     tender.ServiceType,
		Status:          TenderStatusCreated,
		OrganizationId:  tender.OrganizationId,
		CreatorUsername: tender.CreatorUsername,
		Version:         1,
		CreatedAt:       tender.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusCreated, returningTender)
}
