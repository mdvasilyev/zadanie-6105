package tender

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (s *Service) Rollback(db *sql.DB, ctx *gin.Context) {
	ctx.Header("Content-Type", "application/json")

	tenderId := ctx.Param("tenderId")
	if tenderId == "" || len(tenderId) > 100 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid tenderId"})
	}

	newVersion, err := strconv.Atoi(ctx.Param("version"))
	if err != nil || newVersion < 1 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Version must be >= 1"})
		return
	}

	username := ctx.Query("username")
	if username == "" {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Username is required"})
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	queryGet := "SELECT version, creator_username FROM tender WHERE id = $1"

	var currentVersion int
	var creatorUsername string

	err = tx.QueryRowContext(ctx, queryGet, tenderId).Scan(&currentVersion, &creatorUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Post failed: %v, unable to rollback: %v\n", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if newVersion >= currentVersion {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "No such a version. Latest version is " + strconv.Itoa(currentVersion)})
		return
	}

	if username != creatorUsername {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return
	}

	var newTender Tender

	queryGetDiff := "SELECT * FROM tender_diff WHERE id = $1 AND version = $2"

	err = tx.QueryRowContext(ctx, queryGetDiff, tenderId, newVersion).Scan(&newTender.Id, &newTender.Name, &newTender.Description, &newTender.Status, &newTender.ServiceType, &newTender.Version, &newTender.OrganizationId, &newTender.CreatorUsername, &newTender.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Version not found"})
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
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

	queryInsertDiff := "INSERT INTO tender_diff (id, name, description, status, service_type, version, organization_id, creator_username, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	_, err = tx.ExecContext(ctx, queryInsertDiff, newTender.Id, newTender.Name, newTender.Description, newTender.Status, newTender.ServiceType, newTender.Version+1, newTender.OrganizationId, newTender.CreatorUsername, newTender.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Failed to rollback: %v", rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
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
