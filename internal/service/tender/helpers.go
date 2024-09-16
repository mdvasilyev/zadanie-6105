package tender

import (
	"database/sql"
	"errors"
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
	var userExists bool

	checkUserQuery := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`
	err := db.QueryRow(checkUserQuery, username).Scan(&userExists)
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
	queryGet := "SELECT version, creator_username FROM tender WHERE id = $1"

	var currentVersion int
	var creatorUsername string

	err := tx.QueryRowContext(ctx, queryGet, tenderId).Scan(&currentVersion, &creatorUsername)
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
