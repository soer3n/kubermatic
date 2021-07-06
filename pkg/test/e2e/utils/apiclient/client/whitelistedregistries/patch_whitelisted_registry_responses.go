// Code generated by go-swagger; DO NOT EDIT.

package whitelistedregistries

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"k8c.io/kubermatic/v2/pkg/test/e2e/utils/apiclient/models"
)

// PatchWhitelistedRegistryReader is a Reader for the PatchWhitelistedRegistry structure.
type PatchWhitelistedRegistryReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PatchWhitelistedRegistryReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPatchWhitelistedRegistryOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewPatchWhitelistedRegistryUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewPatchWhitelistedRegistryForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewPatchWhitelistedRegistryDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewPatchWhitelistedRegistryOK creates a PatchWhitelistedRegistryOK with default headers values
func NewPatchWhitelistedRegistryOK() *PatchWhitelistedRegistryOK {
	return &PatchWhitelistedRegistryOK{}
}

/*PatchWhitelistedRegistryOK handles this case with default header values.

ConstraintTemplate
*/
type PatchWhitelistedRegistryOK struct {
	Payload *models.ConstraintTemplate
}

func (o *PatchWhitelistedRegistryOK) Error() string {
	return fmt.Sprintf("[PATCH /api/v2/whitelistedregistries/{whitelisted_registry}][%d] patchWhitelistedRegistryOK  %+v", 200, o.Payload)
}

func (o *PatchWhitelistedRegistryOK) GetPayload() *models.ConstraintTemplate {
	return o.Payload
}

func (o *PatchWhitelistedRegistryOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ConstraintTemplate)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPatchWhitelistedRegistryUnauthorized creates a PatchWhitelistedRegistryUnauthorized with default headers values
func NewPatchWhitelistedRegistryUnauthorized() *PatchWhitelistedRegistryUnauthorized {
	return &PatchWhitelistedRegistryUnauthorized{}
}

/*PatchWhitelistedRegistryUnauthorized handles this case with default header values.

EmptyResponse is a empty response
*/
type PatchWhitelistedRegistryUnauthorized struct {
}

func (o *PatchWhitelistedRegistryUnauthorized) Error() string {
	return fmt.Sprintf("[PATCH /api/v2/whitelistedregistries/{whitelisted_registry}][%d] patchWhitelistedRegistryUnauthorized ", 401)
}

func (o *PatchWhitelistedRegistryUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPatchWhitelistedRegistryForbidden creates a PatchWhitelistedRegistryForbidden with default headers values
func NewPatchWhitelistedRegistryForbidden() *PatchWhitelistedRegistryForbidden {
	return &PatchWhitelistedRegistryForbidden{}
}

/*PatchWhitelistedRegistryForbidden handles this case with default header values.

EmptyResponse is a empty response
*/
type PatchWhitelistedRegistryForbidden struct {
}

func (o *PatchWhitelistedRegistryForbidden) Error() string {
	return fmt.Sprintf("[PATCH /api/v2/whitelistedregistries/{whitelisted_registry}][%d] patchWhitelistedRegistryForbidden ", 403)
}

func (o *PatchWhitelistedRegistryForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPatchWhitelistedRegistryDefault creates a PatchWhitelistedRegistryDefault with default headers values
func NewPatchWhitelistedRegistryDefault(code int) *PatchWhitelistedRegistryDefault {
	return &PatchWhitelistedRegistryDefault{
		_statusCode: code,
	}
}

/*PatchWhitelistedRegistryDefault handles this case with default header values.

errorResponse
*/
type PatchWhitelistedRegistryDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the patch whitelisted registry default response
func (o *PatchWhitelistedRegistryDefault) Code() int {
	return o._statusCode
}

func (o *PatchWhitelistedRegistryDefault) Error() string {
	return fmt.Sprintf("[PATCH /api/v2/whitelistedregistries/{whitelisted_registry}][%d] patchWhitelistedRegistry default  %+v", o._statusCode, o.Payload)
}

func (o *PatchWhitelistedRegistryDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *PatchWhitelistedRegistryDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
