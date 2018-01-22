package alexa

// PlainSpeech returns OutputSpeech configured for basic text based output
func PlainSpeech(text string) *OutputSpeech {
	return &OutputSpeech{
		Type: "PlainText",
		Text: text,
	}
}

// SSMLSpeech returns OutputSpeech configured for SSML based output
func SSMLSpeech(ssml string) *OutputSpeech {
	return &OutputSpeech{
		Type: "SSML",
		SSML: ssml,
	}
}
