package tender

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
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
