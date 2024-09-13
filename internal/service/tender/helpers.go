package tender

import "errors"

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
