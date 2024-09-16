package tender

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func validateServiceType(serviceType string) error {
	switch TenderServiceType(serviceType) {
	case TenderServiceTypeDelivery, TenderServiceTypeConstruction, TenderServiceTypeManufacture:
		return nil
	}

	return errors.New("invalid service type")
}

func validateStatus(status string) error {
	switch TenderStatus(status) {
	case TenderStatusCreated, TenderStatusPublished, TenderStatusClosed:
		return nil
	}

	return errors.New("invalid status")
}

func checkUserExistence(db *sql.DB, ctx *gin.Context, username string) bool {
	query := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`

	var userExists bool

	err := db.QueryRow(query, username).Scan(&userExists)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return false
	}
	if !userExists {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{"reason": "Unauthorized user"})
		return false
	}

	return true
}

func extractTenders(ctx *gin.Context, rows *sql.Rows) ([]Tender, bool) {
	var tenders []Tender

	for rows.Next() {
		var t Tender
		err := rows.Scan(&t.Id, &t.Name, &t.Description, &t.Status, &t.ServiceType, &t.Version, &t.OrganizationId, &t.CreatorUsername, &t.CreatedAt)
		if err != nil {
			ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Tenders not found"})
			return nil, false
		}
		tenders = append(tenders, t)
	}

	return tenders, true
}

func checkVersionAndUsername(tx *sql.Tx, ctx *gin.Context, version int, username string, tenderId string) (int, bool) {
	query := "SELECT version, creator_username FROM tender WHERE id = $1"

	var currentVersion int
	var creatorUsername string

	err := tx.QueryRowContext(ctx, query, tenderId).Scan(&currentVersion, &creatorUsername)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return 0, false
	}

	if version >= currentVersion {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"reason": "No such a version. Latest version is " + strconv.Itoa(currentVersion)})
		return 0, false
	}

	if username != creatorUsername {
		ctx.IndentedJSON(http.StatusForbidden, gin.H{"reason": "Wrong username"})
		return 0, false
	}

	return currentVersion, true
}

func insertTender(tx *sql.Tx, ctx *gin.Context, tender Tender) (Tender, bool) {
	query := "INSERT INTO tender (name, description, status, service_type, version, organization_id, creator_username) VALUES ($1, $2, $3, $4, 1, $5, $6) RETURNING id, created_at"

	err := tx.QueryRowContext(ctx, query, tender.Name, tender.Description, TenderStatusCreated, tender.ServiceType, tender.OrganizationId, tender.CreatorUsername).Scan(&tender.Id, &tender.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("err: %v, rollbackErr: %v", err, rollbackErr)})
			return tender, false
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return tender, false
	}

	return tender, true
}

func insertTenderDiff(tx *sql.Tx, ctx *gin.Context, tender Tender) bool {
	query := "INSERT INTO tender_diff (id, name, description, status, service_type, version, organization_id, creator_username, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"

	_, err := tx.ExecContext(ctx, query, tender.Id, tender.Name, tender.Description, tender.Status, tender.ServiceType, tender.Version+1, tender.OrganizationId, tender.CreatorUsername, tender.CreatedAt)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": fmt.Sprintf("Failed to rollback: %v", rollbackErr)})
			return false
		}
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return false
	}

	return true
}

func getTenderById(tx *sql.Tx, ctx *gin.Context, tenderId string) (Tender, bool) {
	query := "SELECT * FROM tender WHERE id = $1"

	var tender Tender

	err := tx.QueryRowContext(ctx, query, tenderId).Scan(&tender.Id, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.OrganizationId, &tender.CreatorUsername, &tender.CreatedAt)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Tender not found"})
		return tender, false
	}

	return tender, true
}

func getTenderByIdAndVersion(tx *sql.Tx, ctx *gin.Context, tenderId string, version int) (Tender, bool) {
	query := "SELECT * FROM tender_diff WHERE id = $1 AND version = $2"

	var tender Tender

	err := tx.QueryRowContext(ctx, query, tenderId, version).Scan(&tender.Id, &tender.Name, &tender.Description, &tender.Status, &tender.ServiceType, &tender.Version, &tender.OrganizationId, &tender.CreatorUsername, &tender.CreatedAt)
	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"reason": "Version not found"})
		return tender, false
	}

	return tender, true
}
