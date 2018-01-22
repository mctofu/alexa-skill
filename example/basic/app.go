package main

import (
	"context"
	"fmt"

	"github.com/mctofu/alexa-skill/alexa"
)

type sessionHandler struct {
}

func (h *sessionHandler) OnLaunch(ctx context.Context, r *alexa.Request, s *alexa.Session, c *alexa.Context) (*alexa.Response, error) {
	return &alexa.Response{
		OutputSpeech: alexa.PlainSpeech("Welcome to my app"),
	}, nil
}

func (h *sessionHandler) OnIntent(ctx context.Context, r *alexa.Request, s *alexa.Session, c *alexa.Context) (*alexa.Response, error) {
	switch r.Intent.Name {
	case "AMAZON.StopIntent":
		return &alexa.Response{
			OutputSpeech:     alexa.PlainSpeech("Goodbye"),
			ShouldSessionEnd: true,
		}, nil
	}

	return nil, fmt.Errorf("unhandled intent: %s", r.Intent.Name)
}

func (h *sessionHandler) OnSessionEnded(ctx context.Context, r *alexa.Request, s *alexa.Session, c *alexa.Context) error {
	return nil
}

func newApp(appID string) *alexa.App {
	return &alexa.App{
		ApplicationID:  appID,
		SessionHandler: &sessionHandler{},
	}
}
