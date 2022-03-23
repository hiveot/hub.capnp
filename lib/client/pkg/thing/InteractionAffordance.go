// Package thing with API interface definitions for the ExposedThing and ConsumedThing classes
package thing

// Form can be viewed as a statement of "To perform an operation type operation on form context, make a
// request method request to submission target" where the optional form fields may further describe the required
// request. In Thing Descriptions, the form context is the surrounding Object, such as Properties, Actions, and
// Events or the Thing itself for meta-interactions.
// (I this isn't clear then you are not alone)
type Form struct {
	Href        string `json:"href"`
	ContentType string `json:"contentType"`

	// operations types of a form as per https://www.w3.org/TR/wot-thing-description11/#form
	// readproperty, writeproperty, ...
	Op string `json:"op"`
}

// InteractionAffordance metadata of a Thing that suggests to Consumers how to interact with the Thing
// This is a DataSchema for the purpose of defining property, actions and events
type InteractionAffordance struct {
	// JSON-LD keyword to label the object with semantic tags (or types)
	AtType string `json:"@type,omitempty"`
	// Provides a human-readable title in the default language
	Title string `json:"title,omitempty"`
	// Provides a multi-language human-readable titles
	Titles []string `json:"titles,omitempty"`
	// Provides additional (human-readable) information based on a default language
	Description string `json:"description,omitempty"`
	// Provides additional nulti-language information
	Descriptions []string `json:"descriptions,omitempty"`

	// Form hypermedia controls to describe how an operation can be performed
	// Forms are serializations of Protocol Bindings.
	Forms []Form `json:"forms"`

	// Define URI template variables according to [RFC6570] as collection based on DataSchema declarations.
	// ... right
	UriVariables map[string]DataSchema `json:"uriVariables,omitempty"`
}
