package resources

import "github.com/OpenNebula/one/src/oca/go/src/goca/errors"

// NoExists indicate if an entity exists in checking the error code returned from an Info call
func NoExists(err error) bool {

	respErr, ok := err.(*errors.ResponseError)

	// expected case, the entity does not exists so we doesn't return an error
	if ok && respErr.Code == errors.OneNoExistsError {
		return true
	}

	return false
}
