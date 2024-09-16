package tender

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) PutStatus(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	tenderId := ctx.Param("tenderId")
	if tenderId == "" || len(tenderId) > 100 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid tenderId"})
		return
	}

	username := ctx.Query("username")
	if username == "" {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	if userExists := checkUserExistence(db, ctx, username); !userExists {
		return
	}

	newStatus := ctx.Query("status")
	if err := validateStatus(newStatus); newStatus == "" || err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid status"})
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	tender, ok := getTenderById(tx, ctx, tenderId)
	if !ok {
		return
	}

	if newStatus == string(tender.Status) {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": fmt.Sprintf("Status is already %v", newStatus)})
		return
	}

	if username != tender.CreatorUsername {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return
	}

	queryUpdate := "UPDATE tender SET status = $1, version = $2 WHERE id = $3"

	_, err = tx.ExecContext(ctx, queryUpdate, newStatus, tender.Version+1, tenderId)
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
		Status:          TenderStatus(newStatus),
		ServiceType:     tender.ServiceType,
		Version:         tender.Version + 1,
		OrganizationId:  tender.OrganizationId,
		CreatorUsername: tender.CreatorUsername,
		CreatedAt:       tender.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusOK, returningTender)
}
