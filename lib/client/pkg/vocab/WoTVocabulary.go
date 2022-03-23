// Package vocab with WoT and JSON-LD defined vocabulary
package vocab

// Core vocabulary definitions
// See https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#sec-core-vocabulary-definition
// Thing, dataschema Vocabulary
const (
	WoTAtType       = "@type"
	WoTAtContext    = "@context"
	WoTAnyURI       = "https://www.w3.org/2019/wot/thing/v1"
	WoTActions      = "actions"
	WoTCreated      = "created"
	WoTDescription  = "description"
	WoTDescriptions = "descriptions"
	WoTEvents       = "events"
	WoTForms        = "forms"
	WoTID           = "id"
	WoTLinks        = "links"
	WoTModified     = "modified"
	WoTProperties   = "properties"
	WoTSecurity     = "security"
	WoTSupport      = "support"
	WoTTitle        = "title"
	WoTTitles       = "titles"
	WoTVersion      = "version"
)

// additional data schema vocab
const (
	WoTConst            = "const"
	WoTDataType         = "type"
	WoTDataTypeAnyURI   = "anyURI" // simple type
	WoTDataTypeArray    = "array"
	WoTDataTypeBool     = "boolean"  // simple type
	WoTDataTypeDateTime = "dateTime" // ISO8601: YYYY-MM-DDTHH:MM:SS.sss[-TZ|+TZ|z]

	WoTDataTypeInteger     = "integer"     // simple type
	WoTDataTypeUnsignedInt = "unsignedInt" // simple type
	WoTDataTypeNumber      = "number"
	WoTDataTypeObject      = "object"
	WoTDataTypeString      = "string" // simple type
	// WoTDouble              = "double" // min, max of number are doubles
	WoTEnum      = "enum"
	WoTFormat    = "format"
	WoTHref      = "href"
	WoTInput     = "input"
	WoTMaximum   = "maximum"
	WoTMaxItems  = "maxItems"
	WoTMaxLength = "maxLength"
	WoTMinimum   = "minimum"
	WoTMinItems  = "minItems"
	WoTMinLength = "minLength"
	WoTOperation = "op"
	WoTOutput    = "output"
	WoTReadOnly  = "readOnly"
	WoTRequired  = "required"
	WoTWriteOnly = "writeOnly"
	WoTUnit      = "unit"
)

// additional security schemas
// Intended for use by Hub services. WoST devices don't need them as they don't run a server
const (
	WoTNoSecurityScheme     = "NoSecurityScheme"
	WoTBasicSecurityScheme  = "BasicSecurityScheme"
	WoTDigestSecurityScheme = "DigestSecurityScheme"
	WoTAPIKeySecurityScheme = "APIKeySecurityScheme"
	WoTBearerSecurityScheme = "BearerSecurityScheme"
	WoTPSKSecurityScheme    = "PSKSecurityScheme"
	WoTOAuth2SecurityScheme = "OAuth2SecurityScheme"
)
