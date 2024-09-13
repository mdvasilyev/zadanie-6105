package bid

import "errors"

func validateStatus(status string) error {
	switch BidStatus(status) {
	case BidStatusCreated, BidStatusPublished, BidStatusCancelled:
		return nil
	}
	return errors.New("invalid status")
}
