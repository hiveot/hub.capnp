// Package api with WoT definitions in golang
// Credits: https://github.com/draggett/arena-webhub
// https://www.w3.org/TR/wot-thing-description/
package td

// ThingDescription defining a thing from the Web of Things
// Based on [draft 24 Nov 2020](https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#introduction)
// Optional parameters are defined with omitempty
type ThingDescription struct {
	Context      []string                 `json:"@context"`         // JSON-LD context: http://www.w3.org/ns/td
	Type         []string                 `json:"@type,omitempty"`  // JSON-LD context
	ID           *string                  `json:"id,omitempty"`     // URI with unique identifier for the thing
	Title        string                   `json:"title"`            // human friendly name for the thing
	Titles       *map[string]string       `json:"titles,omitempty"` // Multi language titles
	Description  *string                  `json:"description,omitempty"`
	Descriptions *map[string]string       `json:"descriptions,omitempty"`
	Version      *VersionInfo             `json:"version,omitempty"`
	Created      *string                  `json:"created,omitempty"`    // DateTime
	Modified     *string                  `json:"modified,omitempty"`   // DateTime
	Support      *string                  `json:"support,omitempty"`    // URI with support contact
	Properties   map[string]ThingProperty `json:"properties,omitempty"` // thing property map
	Actions      map[string]ThingAction   `json:"actions,omitempty"`    // thing action map
	Events       map[string]ThingEvent    `json:"events,omitempty"`     // thing event map
	Links        []string                 `json:"links,omitempty"`
	Forms        []string
	Security     []SecurityScheme `json:"security"` // ???
}

// VersionInfo with version information
type VersionInfo string

// SecurityScheme see: https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#securityscheme
type SecurityScheme struct {
	Type         string             `json:"@type,omitempty"`        // JSON-LD keyword
	Description  *string            `json:"description,omitempty"`  // Human readible info in default language
	Descriptions *map[string]string `json:"descriptions,omitempty"` // Human readible info in different languages
	Proxy        string             `json:"proxy,omitempty"`        // URI of optional proxy server
	Scheme       string             `json:"scheme,omitempty"`       // one of nosec, basic, digest, bearer, psk, oauth2 or apikey
}

// ThingProperty defining a property of a Thing
type ThingProperty struct {
	// thingID string   // ID of the thing this property belongs to
	// Name        string `json:"name"`        // name of the property
	Description string `json:"description"` // description of the property
	Type        string `json:"type"`        // property type
	// Value       string `json:"value"`       // current value of the property
	Minimum int `json:"minimum"` // optional minimum value
	Maximum int `json:"maximum"` // optional maximum value
}

// ThingAction defining an action of a Thing
type ThingAction struct {
	// Title       string `json:"title"`       // human name of the action
	Description string `json:"description"` // description of the action
	Input       struct {
		Type       string                   `json:"type"` // input type: Object
		Properties map[string]ThingProperty `json:"properties"`
	} `json:"input"` // accepted input for the action
	Output struct {
		Type       string                   `json:"type"` // input type: Object
		Properties map[string]ThingProperty `json:"properties"`
	} `json:"output"` // accepted outputs for the action
	Safe       bool `json:"safe,omitempty"` // no internal state involved
	Idempotent bool `json:"idempotennt"`    // Action can be repeated with same result
}

// ThingEvent defining an event of a Thing
type ThingEvent struct {
	// Name string `json:"name"` // name of the event
	Description string `json:"description"` // description of the event
	Data        struct {
		Type string `json:"type"` // type of data
	} `json:"data,omitempty"` // data schema of event ???
	// Subscription string `json:"subscription,omitempty"`  // data to be passed upon subscription ???
	// Cancellation string `json:"canbcellation,omitempty"` // data to pass to cancel a subscription ???
}
