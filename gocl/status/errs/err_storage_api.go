package errs

import "errors"

var ErrStorageAPIMissInContext = errors.New("specified context doesn't provide instance of storage API")
