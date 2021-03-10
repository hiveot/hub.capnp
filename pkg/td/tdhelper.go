package td

// tbd json-ld parsers:
// Most popular; https://github.com/xeipuuv/gojsonschema
// Other:  https://github.com/piprate/json-gold

// CreateTD creates a new Thing Description document
func CreateTD(context string) *ThingDescription {
	td := ThingDescription{
		Context: context,
	}
	return td
}
