package tender

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Rollback(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	tenderId, ok := getTenderId(ctx)
	if !ok {
		return
	}

	newVersion, ok := getVersion(ctx)
	if !ok {
		return
	}

	username, ok := getUsername(ctx)
	if !ok {
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	currentVersion, ok := checkVersionAndUsername(tx, ctx, newVersion, username, tenderId)
	if !ok {
		return
	}

	newTender, ok := getTenderByIdAndVersion(tx, ctx, tenderId, newVersion)
	if !ok {
		return
	}

	queryUpdate := "UPDATE tender SET name = $1, description = $2, status = $3, service_type = $4, version = $5, organization_id = $6, creator_username = $7, created_at = $8 WHERE id = $9"

	_, err = tx.ExecContext(ctx, queryUpdate, newTender.Name, newTender.Description, newTender.Status, newTender.ServiceType, currentVersion+1, newTender.OrganizationId, newTender.CreatorUsername, newTender.CreatedAt, newTender.Id)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Failed to rollback: %v", rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if !insertTenderDiff(tx, ctx, newTender) {
		return
	}

	if err = tx.Commit(); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	returningTender := Tender{
		Id:              newTender.Id,
		Name:            newTender.Name,
		Description:     newTender.Description,
		Status:          newTender.Status,
		ServiceType:     newTender.ServiceType,
		Version:         currentVersion + 1,
		OrganizationId:  newTender.OrganizationId,
		CreatorUsername: newTender.CreatorUsername,
		CreatedAt:       newTender.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusOK, returningTender)
}
