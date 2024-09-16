package tender

import (
	"database/sql"
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

	var tender Tender
	if err = ctx.ShouldBindJSON(&tender); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid request data"})
		return
	}

	tender, ok := insertTender(tx, ctx, tender)
	if !ok {
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
