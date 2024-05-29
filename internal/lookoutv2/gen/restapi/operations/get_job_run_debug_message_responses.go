// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/armadaproject/armada/internal/lookoutv2/gen/models"
)

// GetJobRunDebugMessageOKCode is the HTTP code returned for type GetJobRunDebugMessageOK
const GetJobRunDebugMessageOKCode int = 200

/*
GetJobRunDebugMessageOK Returns debug message for specific job run (if present)

swagger:response getJobRunDebugMessageOK
*/
type GetJobRunDebugMessageOK struct {

	/*
	  In: Body
	*/
	Payload *GetJobRunDebugMessageOKBody `json:"body,omitempty"`
}

// NewGetJobRunDebugMessageOK creates GetJobRunDebugMessageOK with default headers values
func NewGetJobRunDebugMessageOK() *GetJobRunDebugMessageOK {

	return &GetJobRunDebugMessageOK{}
}

// WithPayload adds the payload to the get job run debug message o k response
func (o *GetJobRunDebugMessageOK) WithPayload(payload *GetJobRunDebugMessageOKBody) *GetJobRunDebugMessageOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get job run debug message o k response
func (o *GetJobRunDebugMessageOK) SetPayload(payload *GetJobRunDebugMessageOKBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetJobRunDebugMessageOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetJobRunDebugMessageBadRequestCode is the HTTP code returned for type GetJobRunDebugMessageBadRequest
const GetJobRunDebugMessageBadRequestCode int = 400

/*
GetJobRunDebugMessageBadRequest Error response

swagger:response getJobRunDebugMessageBadRequest
*/
type GetJobRunDebugMessageBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetJobRunDebugMessageBadRequest creates GetJobRunDebugMessageBadRequest with default headers values
func NewGetJobRunDebugMessageBadRequest() *GetJobRunDebugMessageBadRequest {

	return &GetJobRunDebugMessageBadRequest{}
}

// WithPayload adds the payload to the get job run debug message bad request response
func (o *GetJobRunDebugMessageBadRequest) WithPayload(payload *models.Error) *GetJobRunDebugMessageBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get job run debug message bad request response
func (o *GetJobRunDebugMessageBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetJobRunDebugMessageBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*
GetJobRunDebugMessageDefault Error response

swagger:response getJobRunDebugMessageDefault
*/
type GetJobRunDebugMessageDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetJobRunDebugMessageDefault creates GetJobRunDebugMessageDefault with default headers values
func NewGetJobRunDebugMessageDefault(code int) *GetJobRunDebugMessageDefault {
	if code <= 0 {
		code = 500
	}

	return &GetJobRunDebugMessageDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get job run debug message default response
func (o *GetJobRunDebugMessageDefault) WithStatusCode(code int) *GetJobRunDebugMessageDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get job run debug message default response
func (o *GetJobRunDebugMessageDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get job run debug message default response
func (o *GetJobRunDebugMessageDefault) WithPayload(payload *models.Error) *GetJobRunDebugMessageDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get job run debug message default response
func (o *GetJobRunDebugMessageDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetJobRunDebugMessageDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}