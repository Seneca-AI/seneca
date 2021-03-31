package senecaerror

import (
	"errors"
	"fmt"
	"net/http"
)

// WriteErrorToHTTPResponse appropriately writes the given error to the give response.
// If it is a UserError, an external message is printed to the response.
// Params:
//		w http.ResponseWriter: the response to write to
//		err error: the error to handle
// Returns: None.
func WriteErrorToHTTPResponse(w http.ResponseWriter, err error) {
	var ue *UserError
	if errors.As(err, &ue) {
		w.WriteHeader(400)
		fmt.Fprintf(w, fmt.Sprintf("Error: %q", ue.ExternalMessage))
	} else {
		w.WriteHeader(500)
		fmt.Printf("Error: Internal error occurred")
	}
}

// UserError indicates that the error was caused by the user inducing
// some invalid state, like trying to upload two videos with the same create time.
type UserError struct {
	Err             error
	UserID          string
	ExternalMessage string
}

// NewUserError creates a new UserError for the given UserError with the given error message.
func NewUserError(userID string, err error, externalMessage string) *UserError {
	return &UserError{
		Err:             err,
		UserID:          userID,
		ExternalMessage: externalMessage,
	}
}

// Error returns the full error message for a UserError.
func (ue *UserError) Error() string {
	return ue.UserID + ": " + ue.Err.Error()
}

// CloudError is used for errors returned by clou clients.
type CloudError struct {
	Err error
}

// Error returns the full error message for a CloudError.
func (ce *CloudError) Error() string {
	return ce.Err.Error()
}

// NewCloudError creates a new CloudError with the given error.
func NewCloudError(err error) *CloudError {
	return &CloudError{Err: err}
}

// BadStateError indicates that Seneca is in a bad state.
type BadStateError struct {
	Err error
}

// Error returns the full error message for a BadStateError.
func (bse *BadStateError) Error() string {
	return bse.Err.Error()
}

// NewBadStateError returns a new BadStateError.
func NewBadStateError(err error) *BadStateError {
	return &BadStateError{Err: err}
}

// NotFoundError indicates that whatever the caller asked for was not found.
type NotFoundError struct {
	Err error
}

// Error returns the full error message for a NotFoundError.
func (nfe *NotFoundError) Error() string {
	return nfe.Err.Error()
}

// NewNotFoundError returns a new NotFoundError.
func NewNotFoundError(err error) *NotFoundError {
	return &NotFoundError{Err: err}
}
