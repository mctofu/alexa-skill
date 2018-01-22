package alexa

// RequestEnvelope is the data provided by the alexa skill service when initiating a request
type RequestEnvelope struct {
	Version string   `json:"version"`
	Session *Session `json:"session"`
	Request *Request `json:"request"`
	Context *Context `json:"context"`
}

type Request struct {
	Type                 string `json:"type"`
	RequestID            string `json:"requestId"`
	Timestamp            string `json:"timestamp"`
	Intent               Intent `json:"intent"`
	Locale               string `json:"locale"`
	Token                string `json:"token"`
	OffsetInMilliseconds int64  `json:"offsetInMilliseconds"`
}

type Intent struct {
	Name  string                `json:"name"`
	Slots map[string]IntentSlot `json:"slots"`
}

type IntentSlot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Session struct {
	New        bool   `json:"new"`
	SessionID  string `json:"sessionId"`
	Attributes struct {
		String map[string]interface{} `json:"string"`
	} `json:"attributes"`
	User        User               `json:"user"`
	Application RequestApplication `json:"application"`
}

type RequestApplication struct {
	ApplicationID string `json:"applicationId"`
}

type AudioPlayer struct {
	PlayerActivity string `json:"playerActivity"`
}

type SupportedInterfaces struct {
	AudioPlayer AudioPlayer `json:"AudioPlayer"`
}

type Device struct {
	ID         string              `json:"deviceId"`
	Interfaces SupportedInterfaces `json:"supportedInterfaces"`
}

type System struct {
	Application RequestApplication `json:"application"`
	User        User               `json:"user"`
	Device      Device             `json:"device"`
	APIEndpoint string             `json:"apiEndpoint"`
}

type Context struct {
	AudioPlayer AudioPlayer `json:"AudioPlayer"`
	System      System      `json:"System"`
}

type User struct {
	UserID      string      `json:"userId"`
	AccessToken string      `json:"accessToken"`
	Permissions Permissions `json:"permissions"`
}

type Permissions struct {
	ConsentToken string `json:"consentToken,omitempty"`
}

// ResponseEnvelope is data the alexa skill service expects in response to a request
type ResponseEnvelope struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          *Response              `json:"response"`
}

type Response struct {
	OutputSpeech     *OutputSpeech `json:"outputSpeech,omitempty"`
	Card             *Card         `json:"card,omitempty"`
	Reprompt         *Reprompt     `json:"reprompt,omitempty"`
	Directives       *[]Directive  `json:"directives,omitempty"`
	ShouldSessionEnd bool          `json:"shouldEndSession"`
}

// OutputSpeech contains the data the defines what Alexa should say to the user.
type OutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

type Card struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

type Image struct {
	SmallImageURL string `json:"smallImageUrl,omitempty"`
	LargeImageURL string `json:"largeImageUrl,omitempty"`
}

type Reprompt struct {
	OutputSpeech *OutputSpeech `json:"outputSpeech,omitempty"`
}

type Directive struct {
	Type         string `json:"type"`
	PlayBehavior string `json:"playBehavior,omitempty"`
	AudioItem    *struct {
		Stream *Stream `json:"stream,omitempty"`
	} `json:"audioItem,omitempty"`
}

type Stream struct {
	Token                string `json:"token"`
	URL                  string `json:"url"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
}
