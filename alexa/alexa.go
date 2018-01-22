package alexa

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

const sdkVersion = "1.0"

// AppHandler responds to requests from the alexa skill service
type AppHandler interface {
	Handle(ctx context.Context, r *RequestEnvelope) (*ResponseEnvelope, error)
}

// SessionHandler handles the main Launch, Intent and SessionEnded requests
type SessionHandler interface {
	OnLaunch(context.Context, *Request, *Session, *Context) (*Response, error)
	OnIntent(context.Context, *Request, *Session, *Context) (*Response, error)
	OnSessionEnded(context.Context, *Request, *Session, *Context) error
}

// AudioHandler handles AudioPlayer requests
type AudioHandler interface {
	OnPlaybackStarted(context.Context, *Request, *Context) ([]Directive, error)
	OnPlaybackFinished(context.Context, *Request, *Context) ([]Directive, error)
	OnPlaybackStopped(context.Context, *Request, *Context) error
	OnPlaybackNearlyFinished(context.Context, *Request, *Context) ([]Directive, error)
	OnPlaybackFailed(context.Context, *Request, *Context) ([]Directive, error)
}

// App responds to Alexa requests by routing them to the correct handler
type App struct {
	ApplicationID  string
	SessionHandler SessionHandler
	AudioHandler   AudioHandler
}

// Handle inspects the request type and forwards the request to the appropriate handler
func (a *App) Handle(ctx context.Context, r *RequestEnvelope) (*ResponseEnvelope, error) {
	var err error
	var resp *Response
	switch r.Request.Type {
	case "LaunchRequest":
		resp, err = a.SessionHandler.OnLaunch(ctx, r.Request, r.Session, r.Context)
	case "IntentRequest":
		resp, err = a.SessionHandler.OnIntent(ctx, r.Request, r.Session, r.Context)
	case "SessionEndedRequest":
		err = a.SessionHandler.OnSessionEnded(ctx, r.Request, r.Session, r.Context)
	default:
		switch {
		case strings.HasPrefix(r.Request.Type, "AudioPlayer."):
			resp, err = a.handleAudio(ctx, r)
		default:
			err = ValidationError{"unhandled request type: " + r.Request.Type}
		}
	}
	if err != nil {
		if _, ok := err.(ValidationError); ok {
			return nil, err
		}
		return nil, fmt.Errorf("failed to handle %s request: %v", r.Request.Type, err)
	}

	respBody := &ResponseEnvelope{Version: sdkVersion, Response: resp}
	if r.Session != nil {
		respBody.SessionAttributes = r.Session.Attributes.String
	}

	return respBody, nil
}

func (a *App) handleAudio(ctx context.Context, r *RequestEnvelope) (*Response, error) {
	var directives []Directive
	var err error
	switch r.Request.Type {
	case "AudioPlayer.PlaybackStarted":
		directives, err = a.AudioHandler.OnPlaybackStarted(ctx, r.Request, r.Context)
	case "AudioPlayer.PlaybackFinished":
		directives, err = a.AudioHandler.OnPlaybackFinished(ctx, r.Request, r.Context)
	case "AudioPlayer.PlaybackStopped":
		err = a.AudioHandler.OnPlaybackStopped(ctx, r.Request, r.Context)
	case "AudioPlayer.PlaybackNearlyFinished":
		directives, err = a.AudioHandler.OnPlaybackNearlyFinished(ctx, r.Request, r.Context)
	case "AudioPlayer.PlaybackFailed":
		directives, err = a.AudioHandler.OnPlaybackFailed(ctx, r.Request, r.Context)
	default:
		err = ValidationError{"unhandled request type: " + r.Request.Type}
	}
	if err != nil {
		return nil, err
	}

	resp := &Response{ShouldSessionEnd: true}
	if len(directives) > 0 {
		resp.Directives = &directives
	}

	return resp, nil
}

// ValidationError is an error indicating there was an issue with the request contents.
type ValidationError struct {
	Message string
}

func (v ValidationError) Error() string {
	return v.Message
}

// StrictVerificationApp performs additional verification of the request as specified here:
// https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/developing-an-alexa-skill-as-a-web-service#timestamp
type StrictVerificationApp struct {
	App App
}

// Handle verifies the request before handing it off to the wrapped App
func (s *StrictVerificationApp) Handle(ctx context.Context, r *RequestEnvelope) (*ResponseEnvelope, error) {
	err := validateApplicationID(r, s.App.ApplicationID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	err = validateTimestamp(r, now, float64(150))
	if err != nil {
		return nil, err
	}

	return s.App.Handle(ctx, r)
}

func validateApplicationID(r *RequestEnvelope, applicationID string) error {
	reqAppID := r.Session.Application.ApplicationID
	if reqAppID == "" {
		return ValidationError{"request applicationID missing"}
	}
	if reqAppID != applicationID {
		return ValidationError{fmt.Sprintf("request applicationID mismatch: %s", reqAppID)}
	}

	return nil
}

func validateTimestamp(r *RequestEnvelope, now time.Time, maxDiffSeconds float64) error {
	timestamp, err := time.Parse(time.RFC3339, r.Request.Timestamp)
	if err != nil {
		return ValidationError{fmt.Sprintf("invalid timestamp %s: %v", r.Request.Timestamp, err)}
	}

	diff := now.Sub(timestamp).Seconds()
	if math.Abs(diff) > maxDiffSeconds {
		return ValidationError{fmt.Sprintf("timestamp %s difference of %fs exceeds %fs", timestamp, diff, maxDiffSeconds)}
	}
	return nil
}
