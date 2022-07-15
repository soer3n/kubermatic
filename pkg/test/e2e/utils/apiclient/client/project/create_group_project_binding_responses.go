// Code generated by go-swagger; DO NOT EDIT.

package project

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"k8c.io/kubermatic/v2/pkg/test/e2e/utils/apiclient/models"
)

// CreateGroupProjectBindingReader is a Reader for the CreateGroupProjectBinding structure.
type CreateGroupProjectBindingReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *CreateGroupProjectBindingReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 201:
		result := NewCreateGroupProjectBindingCreated()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewCreateGroupProjectBindingUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewCreateGroupProjectBindingForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewCreateGroupProjectBindingDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewCreateGroupProjectBindingCreated creates a CreateGroupProjectBindingCreated with default headers values
func NewCreateGroupProjectBindingCreated() *CreateGroupProjectBindingCreated {
	return &CreateGroupProjectBindingCreated{}
}

/* CreateGroupProjectBindingCreated describes a response with status code 201, with default header values.

GroupProjectBinding
*/
type CreateGroupProjectBindingCreated struct {
	Payload *models.GroupProjectBinding
}

func (o *CreateGroupProjectBindingCreated) Error() string {
	return fmt.Sprintf("[POST /api/v2/projects/{project_id}/groupbindings][%d] createGroupProjectBindingCreated  %+v", 201, o.Payload)
}
func (o *CreateGroupProjectBindingCreated) GetPayload() *models.GroupProjectBinding {
	return o.Payload
}

func (o *CreateGroupProjectBindingCreated) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GroupProjectBinding)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewCreateGroupProjectBindingUnauthorized creates a CreateGroupProjectBindingUnauthorized with default headers values
func NewCreateGroupProjectBindingUnauthorized() *CreateGroupProjectBindingUnauthorized {
	return &CreateGroupProjectBindingUnauthorized{}
}

/* CreateGroupProjectBindingUnauthorized describes a response with status code 401, with default header values.

EmptyResponse is a empty response
*/
type CreateGroupProjectBindingUnauthorized struct {
}

func (o *CreateGroupProjectBindingUnauthorized) Error() string {
	return fmt.Sprintf("[POST /api/v2/projects/{project_id}/groupbindings][%d] createGroupProjectBindingUnauthorized ", 401)
}

func (o *CreateGroupProjectBindingUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewCreateGroupProjectBindingForbidden creates a CreateGroupProjectBindingForbidden with default headers values
func NewCreateGroupProjectBindingForbidden() *CreateGroupProjectBindingForbidden {
	return &CreateGroupProjectBindingForbidden{}
}

/* CreateGroupProjectBindingForbidden describes a response with status code 403, with default header values.

EmptyResponse is a empty response
*/
type CreateGroupProjectBindingForbidden struct {
}

func (o *CreateGroupProjectBindingForbidden) Error() string {
	return fmt.Sprintf("[POST /api/v2/projects/{project_id}/groupbindings][%d] createGroupProjectBindingForbidden ", 403)
}

func (o *CreateGroupProjectBindingForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewCreateGroupProjectBindingDefault creates a CreateGroupProjectBindingDefault with default headers values
func NewCreateGroupProjectBindingDefault(code int) *CreateGroupProjectBindingDefault {
	return &CreateGroupProjectBindingDefault{
		_statusCode: code,
	}
}

/* CreateGroupProjectBindingDefault describes a response with status code -1, with default header values.

errorResponse
*/
type CreateGroupProjectBindingDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the create group project binding default response
func (o *CreateGroupProjectBindingDefault) Code() int {
	return o._statusCode
}

func (o *CreateGroupProjectBindingDefault) Error() string {
	return fmt.Sprintf("[POST /api/v2/projects/{project_id}/groupbindings][%d] createGroupProjectBinding default  %+v", o._statusCode, o.Payload)
}
func (o *CreateGroupProjectBindingDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *CreateGroupProjectBindingDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}