# W3C Web Of Things vocabulary
# Core vocabulary definitions
# See https://www.w3.org/TR/2020/WD-wot-thing-description11-20201124/#sec-core-vocabulary-definition
# Thing, dataschema Vocabulary
@0xa2483ae53bb8fb47;

using Go = import "/go.capnp";
$Go.package("vocab");
$Go.import("github.com/hiveot/hub/api/go/hubapi");


const woTAtType        :Text = "@type";
const woTAtContext     :Text = "@context";
const woTAnyURI        :Text = "https://www.w3.org/2019/wot/thing/v1";
const woTActions       :Text = "actions";
const woTCreated       :Text = "created";
const woTDescription   :Text = "description";
const woTDescriptions  :Text = "descriptions";
const woTEvents        :Text = "events";
const woTForms         :Text = "forms";
const woTID            :Text = "id";
const woTLinks         :Text = "links";
const woTModified      :Text = "modified";
const woTProperties    :Text = "properties";
const woTSecurity      :Text = "security";
const woTSupport       :Text = "support";
const woTTitle         :Text = "title";
const woTTitles        :Text = "titles";
const woTVersion       :Text = "version";


# additional data schema vocab
const woTConst             :Text = "const";
const woTDataType          :Text = "type";
const woTDataTypeAnyURI    :Text = "anyURI"; # simple type
const woTDataTypeArray     :Text = "array";
const woTDataTypeBool      :Text = "boolean";  # simple type
const woTDataTypeDateTime  :Text = "dateTime"; # ISO8601: YYYY-MM-DDTHH:MM:SS.sss[-TZ|+TZ|z]

const woTDataTypeInteger      :Text = "integer";     # simple type
const woTDataTypeUnsignedInt  :Text = "unsignedInt"; # simple type
const woTDataTypeNumber       :Text = "number";
const woTDataTypeObject       :Text = "object";
const woTDataTypeString       :Text = "string"; # simple type

const woTDataTypeNone        :Text = ""; # no data

	# WoTDouble               :Text = "double"; # min, max of number are doubles
const woTEnum       :Text = "enum";
const woTFormat     :Text = "format";
const woTHref       :Text = "href";
const woTInput      :Text = "input";
const woTMaximum    :Text = "maximum";
const woTMaxItems   :Text = "maxItems";
const woTMaxLength  :Text = "maxLength";
const woTMinimum    :Text = "minimum";
const woTMinItems   :Text = "minItems";
const woTMinLength  :Text = "minLength";
const woTOperation  :Text = "op";
const woTOutput     :Text = "output";
const woTReadOnly   :Text = "readOnly";
const woTRequired   :Text = "required";
const woTWriteOnly  :Text = "writeOnly";
const woTUnit       :Text = "unit";


# additional security schemas
# Intended for use by Hub services. HiveOT devices don't need them as they don't run a server
const woTNoSecurityScheme      :Text = "NoSecurityScheme";
const woTBasicSecurityScheme   :Text = "BasicSecurityScheme";
const woTDigestSecurityScheme  :Text = "DigestSecurityScheme";
const woTAPIKeySecurityScheme  :Text = "APIKeySecurityScheme";
const woTBearerSecurityScheme  :Text = "BearerSecurityScheme";
const woTPSKSecurityScheme     :Text = "PSKSecurityScheme";
const woTOAuth2SecurityScheme  :Text = "OAuth2SecurityScheme";
