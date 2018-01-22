package lambda

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/mctofu/alexa-skill/alexa"
)

// Handler handles an alexa skill lambda request.
type Handler func(ctx context.Context, r *alexa.RequestEnvelope) (*alexa.ResponseEnvelope, error)

// NewAppHandler returns a handler that services requests using the provided AppHandler
func NewAppHandler(app alexa.AppHandler) Handler {
	return app.Handle
}

// NewDebugHandler wraps a handler and provides logging of the request and response
func NewDebugHandler(h Handler) Handler {
	return func(ctx context.Context, r *alexa.RequestEnvelope) (*alexa.ResponseEnvelope, error) {
		var rID string
		lCtx, ok := lambdacontext.FromContext(ctx)
		if ok {
			rID = lCtx.AwsRequestID
		} else {
			rID = "n/a"
		}

		eventJSON, err := json.Marshal(r)
		if err != nil {
			log.Printf("[%s] Request: Failed to marshal: %v\n", rID, err)
		} else {
			log.Printf("[%s] Request: %s\n", rID, eventJSON)
		}

		resp, err := h(ctx, r)
		if err != nil {
			log.Printf("[%s] Error: %v\n", rID, err)
		}

		out, err := json.Marshal(resp)
		if err != nil {
			log.Printf("[%s] Response: Failed to marshal: %v\n", rID, err)
		} else {
			log.Printf("[%s] Response: %s\n", rID, out)
		}
		return resp, err
	}
}
