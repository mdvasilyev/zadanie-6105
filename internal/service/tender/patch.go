package tender

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *Service) Patch(db *sql.DB, ctx *gin.Context) {
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

	var tenderPatch TenderPatch
	if err := ctx.ShouldBindJSON(&tenderPatch); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "Invalid request body"})
		return
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	tender := Tender{}

	queryGet := "SELECT * FROM tender WHERE id = $1"

	err = tx.QueryRowContext(ctx, queryGet, tenderId).Scan(&tender.Id, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.OrganizationId, &tender.CreatorUsername, &tender.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err})
		return
	}

	if username != tender.CreatorUsername {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return
	}

	changes := make(map[string]interface{})
	tenderDiffValues := make(map[string]interface{})

	if tenderPatch.Name != "" {
		changes["name"] = tenderPatch.Name
		tenderDiffValues["name"] = tenderPatch.Name
		tender.Name = tenderPatch.Name
	}

	if tenderPatch.Description != "" {
		changes["description"] = tenderPatch.Description
		tenderDiffValues["description"] = tenderPatch.Description
		tender.Description = tenderPatch.Description
	}

	if tenderPatch.ServiceType != "" {
		changes["service_type"] = tenderPatch.ServiceType
		tenderDiffValues["service_type"] = tenderPatch.ServiceType
		tender.ServiceType = tenderPatch.ServiceType
	}

	changes["version"] = tender.Version + 1

	var updates []string
	var params []interface{}
	paramCounter := 1

	for column, value := range changes {
		updates = append(updates, fmt.Sprintf("%s = $%d", column, paramCounter))
		params = append(params, value)
		paramCounter++
	}

	queryUpdate := fmt.Sprintf("UPDATE tender SET %s WHERE id = $%d", strings.Join(updates, ", "), paramCounter)
	queryDiff := "INSERT INTO tender_diff (id, name, description, status, service_type, version, organization_id, creator_username, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	params = append(params, tenderId)

	_, err = tx.ExecContext(ctx, queryUpdate, params...)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	_, err = tx.ExecContext(ctx, queryDiff, tender.Id, tender.Name, tender.Description, tender.Status, tender.ServiceType, tender.Version+1, tender.OrganizationId, tender.CreatorUsername, tender.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
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
		Id:              tender.Id,
		Name:            tender.Name,
		Description:     tender.Description,
		Status:          tender.Status,
		ServiceType:     tender.ServiceType,
		Version:         tender.Version + 1,
		OrganizationId:  tender.OrganizationId,
		CreatorUsername: tender.CreatorUsername,
		CreatedAt:       tender.CreatedAt,
	}

	ctx.IndentedJSON(http.StatusOK, returningTender)
}
