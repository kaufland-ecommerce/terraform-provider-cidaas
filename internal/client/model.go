package client

type TenantInfo struct {
	CustomFieldFlatten bool   `json:"Custom_field_flatten"`
	TenantKey          string `json:"tenant_key"`
	TenantName         string `json:"tenant_name"`
	VersionInfo        string `json:"versionInfo"`
}

type Hook struct {
	Id            string            `json:"_id,omitempty"`
	AuthType      string            `json:"auth_type,omitempty"`
	Events        []string          `json:"events"`
	URL           string            `json:"url"`
	CreatedTime   string            `json:"createdTime,omitempty"`
	UpdatedTime   string            `json:"updatedTime,omitempty"`
	ApiKeyDetails HookApiKeyDetails `json:"apikeyDetails,omitempty"`
}

type HookApiKeyDetails struct {
	APIKeyPlacement   string `json:"apikey_placement,omitempty"`
	APIKey            string `json:"apikey,omitempty"`
	APIKeyPlaceholder string `json:"apikey_placeholder,omitempty"`
}

type SocialProvider struct {
	Id           string `json:"id,omitempty"`
	SocialId     string `json:"social_id"`
	Name         string `json:"name"`
	ProviderName string `json:"provider_name"`
	ProviderType string `json:"provider_type,omitempty"`
}

type ConsentInstance struct {
	ID          string `json:"id"`
	ConsentName string `json:"consent_name"`
}

type PasswordPolicy struct {
	ID                string `json:"id"`
	PolicyName        string `json:"policy_name"`
	MinimumLength     int64  `json:"minimumLength"`
	NoOfDigits        int64  `json:"noOfDigits"`
	LowerAndUpperCase bool   `json:"lowerAndUpperCase"`
	NoOfSpecialChars  int64  `json:"noOfSpecialChars"`
}

type HostedPage struct {
	ID      string `json:"hosted_page_id" tfsdk:"id"`
	Content string `json:"content" tfsdk:"content"`
	Locale  string `json:"locale" tfsdk:"locale"`
	Url     string `json:"url" tfsdk:"url"`
}

type HostedPageGroup struct {
	ID            string       `json:"_id"`
	CreatedTime   string       `json:"createdTime,omitempty"`
	UpdatedTime   string       `json:"updatedTime,omitempty"`
	DefaultLocale string       `json:"default_locale"`
	GroupOwner    string       `json:"groupOwner"`
	HostedPages   []HostedPage `json:"hosted_pages"`
}

type AppKey struct {
	ID         string `json:"id"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type AllowedGroup struct {
	GroupId      string   `json:"groupId" tfsdk:"group_id"`
	Roles        []string `json:"roles" tfsdk:"roles"`
	DefaultRoles []string `json:"default_roles" tfsdk:"default_roles"`
}

type App struct {
	ID                              string  `json:"id,omitempty"`
	AcceptRolesInTheRegistration    bool    `json:"accept_roles_in_the_registration"`
	ClientId                        string  `json:"client_id,omitempty"`
	ClientSecret                    string  `json:"client_secret,omitempty"`
	ClientName                      string  `json:"client_name"`
	ClientDisplayName               string  `json:"client_display_name"`
	IsRememberMeSelected            bool    `json:"is_remember_me_selected"`
	ClientType                      string  `json:"client_type"`
	AllowDisposableEmail            bool    `json:"allow_disposable_email"`
	FdsEnabled                      bool    `json:"fds_enabled"`
	EnablePasswordlessAuth          bool    `json:"enable_passwordless_auth"`
	EnableDeduplication             bool    `json:"enable_deduplication"`
	CommunicationMediumVerification string  `json:"communication_medium_verification"`
	HostedPageGroup                 string  `json:"hosted_page_group"`
	PrimaryColor                    string  `json:"primaryColor"`
	AccentColor                     string  `json:"accentColor"`
	AutoLoginAfterRegister          bool    `json:"auto_login_after_register"`
	CompanyName                     string  `json:"company_name"`
	CompanyAddress                  string  `json:"company_address"`
	CompanyWebsite                  string  `json:"company_website"`
	TemplateGroupId                 *string `json:"template_group_id"`
	TokenLifetimeInSeconds          int64   `json:"token_lifetime_in_seconds"`
	IdTokenLifetimeInSeconds        int64   `json:"id_token_lifetime_in_seconds"`
	RefreshTokenLifetimeInSeconds   int64   `json:"refresh_token_lifetime_in_seconds"`
	EnableBotDetection              bool    `json:"enable_bot_detection"`
	IsLoginSuccessPageEnabled       bool    `json:"is_login_success_page_enabled"`
	AllowGuestLogin                 bool    `json:"allow_guest_login"`
	JweEnabled                      bool    `json:"jwe_enabled"`
	AlwaysAskMfa                    bool    `json:"always_ask_mfa"`
	PasswordPolicy                  *string `json:"password_policy_ref,omitempty"`
	RegisterWithLoginInformation    bool    `json:"register_with_login_information"`
	AppOwner                        string  `json:"app_owner,omitempty"`
	BotProvider                     string  `json:"bot_provider,omitempty"`

	AppKey *AppKey `json:"appKey,omitempty"`

	AllowLoginWith               []string         `json:"allow_login_with"`
	OperationsAllowedGroups      []AllowedGroup   `json:"operations_allowed_groups"`
	AllowedGroups                []AllowedGroup   `json:"allowed_groups"`
	RedirectUris                 []string         `json:"redirect_uris"`
	AllowedLogoutUrls            []string         `json:"allowed_logout_urls"`
	AllowedScopes                []string         `json:"allowed_scopes"`
	SocialProviders              []SocialProvider `json:"social_providers"`
	CustomProviders              []CustomProvider `json:"custom_providers"`
	AdditionalAccessTokenPayload []string         `json:"additional_access_token_payload"`
	AllowedFields                []string         `json:"allowed_fields"`
	RequiredFields               []string         `json:"required_fields"`
	ConsentRefs                  []string         `json:"consent_refs"`
	ResponseTypes                []string         `json:"response_types"`
	GrantTypes                   []string         `json:"grant_types"`
	AllowedWebOrigins            []string         `json:"allowed_web_origins"`
	AllowedOrigins               []string         `json:"allowed_origins"`
	AllowedMfa                   []string         `json:"allowed_mfa"`
	AllowedRoles                 []string         `json:"allowed_roles"`
}

type RegistrationField struct {
	Internal        bool            `json:"internal"`
	ReadOnly        bool            `json:"readOnly"`
	Claimable       bool            `json:"claimable"`
	Required        bool            `json:"required"`
	Scopes          []string        `json:"scopes"`
	Enabled         bool            `json:"enabled"`
	LocaleText      LocaleText      `json:"localeText"`
	IsGroup         bool            `json:"is_group"`
	IsList          bool            `json:"is_list"`
	ParentGroupID   string          `json:"parent_group_id"`
	FieldType       string          `json:"fieldType"`
	ConsentRefs     []string        `json:"consent_refs"`
	ID              *string         `json:"_id,omitempty"`
	FieldKey        string          `json:"fieldKey"`
	DataType        string          `json:"dataType"`
	Order           int64           `json:"order"`
	FieldDefinition FieldDefinition `json:"fieldDefinition"`
	BaseDataType    string          `json:"baseDataType"`
}
type ConsentLabel struct {
	Label     string `json:"label"`
	LabelText string `json:"label_text"`
}
type LocaleText struct {
	Locale       string       `json:"locale"`
	Language     string       `json:"language"`
	ConsentLabel ConsentLabel `json:"consentLabel"`
}
type FieldDefinition struct {
	Language string `json:"language"`
	Locale   string `json:"locale"`
}

type EmailSenderConfig struct {
	CommunicationMethod string `json:"communicationMethod" tfsdk:"communication_method"`
	ServiceSetupId      string `json:"serviceSetupId,omitempty" tfsdk:"service_setup_id"`
	SenderName          string `json:"senderName" tfsdk:"sender_name"`
	SenderAddress       string `json:"senderAddress" tfsdk:"sender_address"`
}

type SmsSenderConfig struct {
	CommunicationMethod string `json:"communicationMethod" tfsdk:"communication_method"`
	ServiceSetupId      string `json:"serviceSetupId,omitempty" tfsdk:"service_setup_id"`
	SenderName          string `json:"senderName" tfsdk:"sender_name"`
	SenderAddress       string `json:"senderAddress" tfsdk:"sender_address"`
}

type IVRSenderConfig struct {
	CommunicationMethod string   `json:"communicationMethod" tfsdk:"communication_method"`
	ServiceSetupId      string   `json:"serviceSetupId,omitempty" tfsdk:"service_setup_id"`
	SenderAddress       []string `json:"senderAddress" tfsdk:"sender_address"`
}

type PushSenderConfig struct {
	CommunicationMethod string `json:"communicationMethod" tfsdk:"communication_method"`
	ServiceSetupId      string `json:"serviceSetupId,omitempty" tfsdk:"service_setup_id"`
}

type TemplateGroupComSettings struct {
	EmailSenderConfig EmailSenderConfig `json:"email"`
	SmsSenderConfig   SmsSenderConfig   `json:"sms"`
	IVRSenderConfig   IVRSenderConfig   `json:"ivr"`
	PushSenderConfig  PushSenderConfig  `json:"push"`
}

type TemplateGroup struct {
	Id            string                   `json:"id,omitempty"`
	Description   string                   `json:"description"`
	CommSettings  TemplateGroupComSettings `json:"commSettings"`
	DefaultLocale string                   `json:"defaultLocale"`
}

type Template struct {
	ID                  *string `json:"id,omitempty"`
	LastSeededBy        *string `json:"lastSeededBy,omitempty"`
	GroupId             string  `json:"groupId"`
	TemplateKey         string  `json:"templateKey"`
	CommunicationMethod string  `json:"communicationMethod"`
	ProcessingType      string  `json:"processingType"`
	Locale              string  `json:"locale"`
	MessageFormat       string  `json:"messageFormat"`
	Enabled             bool    `json:"enabled"`
	Subject             string  `json:"subject"`
	Content             string  `json:"content"`
}

type CustomProvider struct {
	DisplayName  string `json:"display_name"`
	ProviderName string `json:"provider_name"`
}
