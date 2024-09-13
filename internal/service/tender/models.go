package tender

import (
	"github.com/google/uuid"
	"time"
)

type TenderStatus string
type TenderServiceType string

const (
	TenderStatusCreated   TenderStatus = "Created"
	TenderStatusPublished TenderStatus = "Published"
	TenderStatusClosed    TenderStatus = "Closed"
)

const (
	TenderServiceTypeConstruction TenderServiceType = "Construction"
	TenderServiceTypeDelivery     TenderServiceType = "Delivery"
	TenderServiceTypeManufacture  TenderServiceType = "Manufacture"
)

type Tender struct {
	Id              uuid.UUID         `json:"id"`
	Name            string            `json:"name" binding:"required"`
	Description     string            `json:"description" binding:"required"`
	Status          TenderStatus      `json:"status"`
	ServiceType     TenderServiceType `json:"serviceType" binding:"required"`
	Version         int               `json:"version"`
	OrganizationId  uuid.UUID         `json:"organizationId" binding:"required"`
	CreatorUsername string            `json:"creatorUsername" binding:"required"`
	CreatedAt       time.Time         `json:"createdAt"`
}

type TenderPatch struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ServiceType TenderServiceType `json:"serviceType"`
}
