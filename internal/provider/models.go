package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type TenantInfo struct {
	CustomFieldFlatten types.Bool   `tfsdk:"custom_field_flatten"`
	TenantKey          types.String `tfsdk:"tenant_key"`
	TenantName         types.String `tfsdk:"tenant_name"`
	VersionInfo        types.String `tfsdk:"version_info"`
}

type Hook struct {
	ID            types.String      `tfsdk:"id"`
	LastUpdate    types.String      `tfsdk:"last_updated"`
	Url           string            `tfsdk:"url"`
	AuthType      string            `tfsdk:"auth_type"`
	Events        []string          `tfsdk:"events"`
	APIKeyDetails HookAPIKeyDetails `tfsdk:"apikey_details"`
}

type HookAPIKeyDetails struct {
	APIKey            string `tfsdk:"apikey"`
	APIKeyPlaceholder string `tfsdk:"apikey_placeholder"`
	APIKeyPlacement   string `tfsdk:"apikey_placement"`
}

type SocialProvider struct {
	SocialId     types.String `tfsdk:"social_id"`
	ProviderName types.String `tfsdk:"provider_name"`
	ProviderType types.String `tfsdk:"provider_type"`
	Name         types.String `tfsdk:"name"`
}

type ConsentInstance struct {
	ID          types.String `tfsdk:"id"`
	ConsentName types.String `tfsdk:"consent_name"`
}

type PasswordPolicy struct {
	ID                types.String `tfsdk:"id"`
	PolicyName        types.String `tfsdk:"policy_name"`
	MinimumLength     types.Int64  `tfsdk:"minimum_length"`
	NoOfDigits        types.Int64  `tfsdk:"no_of_digits"`
	LowerAndUpperCase types.Bool   `tfsdk:"lower_and_upper_case"`
	NoOfSpecialChars  types.Int64  `tfsdk:"no_of_special_chars"`
}

type HostedPageGroup struct {
	Name  types.String `tfsdk:"name"`
	Pages types.Map    `tfsdk:"pages"`
}

type App struct {
	ID                               types.String `tfsdk:"id"`
	ClientId                         types.String `tfsdk:"client_id"`
	ClientSecret                     types.String `tfsdk:"client_secret"`
	ClientName                       types.String `tfsdk:"client_name"`
	ClientDisplayName                types.String `tfsdk:"client_display_name"`
	ClientType                       types.String `tfsdk:"client_type"`
	IsRememberMeSelected             types.Bool   `tfsdk:"is_remember_me_selected"`
	AllowDisposableEmail             types.Bool   `tfsdk:"allow_disposable_email"`
	FdsEnabled                       types.Bool   `tfsdk:"fds_enabled"`
	EnablePasswordlessAuth           types.Bool   `tfsdk:"enable_passwordless_auth"`
	EnableDeduplication              types.Bool   `tfsdk:"enable_deduplication"`
	MobileNumberVerificationRequired types.Bool   `tfsdk:"mobile_number_verification_required"`
	HostedPageGroup                  types.String `tfsdk:"hosted_page_group"`
	PrimaryColor                     types.String `tfsdk:"primary_color"`
	AccentColor                      types.String `tfsdk:"accent_color"`
	AutoLoginAfterRegister           types.Bool   `tfsdk:"auto_login_after_register"`
	CompanyName                      types.String `tfsdk:"company_name"`
	CompanyAddress                   types.String `tfsdk:"company_address"`
	CompanyWebsite                   types.String `tfsdk:"company_website"`
	TokenLifetimeInSeconds           types.Int64  `tfsdk:"token_lifetime_in_seconds"`
	IdTokenLifetimeInSeconds         types.Int64  `tfsdk:"id_token_lifetime_in_seconds"`
	RefreshTokenLifetimeInSeconds    types.Int64  `tfsdk:"refresh_token_lifetime_in_seconds"`
	EmailVerificationRequired        types.Bool   `tfsdk:"email_verification_required"`
	EnableBotDetection               types.Bool   `tfsdk:"enable_bot_detection"`
	IsLoginSuccessPageEnabled        types.Bool   `tfsdk:"is_login_success_page_enabled"`
	AllowGuestLogin                  types.Bool   `tfsdk:"allow_guest_login"`
	JweEnabled                       types.Bool   `tfsdk:"jwe_enabled"`
	AlwaysAskMfa                     types.Bool   `tfsdk:"always_ask_mfa"`
	PasswordPolicy                   types.String `tfsdk:"password_policy"`
	RegisterWithLoginInformation     types.Bool   `tfsdk:"register_with_login_information"`
	AppKey                           types.Object `tfsdk:"app_key"`
	TemplateGroupId                  types.String `tfsdk:"template_group_id"`

	AllowedGroups                types.List       `tfsdk:"allowed_groups"`
	OperationsAllowedGroups      types.List       `tfsdk:"operations_allowed_groups"`
	AllowLoginWith               []string         `tfsdk:"allow_login_with"`
	RedirectUris                 []string         `tfsdk:"redirect_uris"`
	AllowedLogoutUrls            []string         `tfsdk:"allowed_logout_urls"`
	SocialProviders              []SocialProvider `tfsdk:"social_providers"`
	AdditionalAccessTokenPayload []string         `tfsdk:"additional_access_token_payload"`
	Scopes                       []string         `tfsdk:"allowed_scopes"`
	AllowedFields                []string         `tfsdk:"allowed_fields"`
	RequiredFields               []string         `tfsdk:"required_fields"`
	ConsentRefs                  []string         `tfsdk:"consent_refs"`
	ResponseTypes                []string         `tfsdk:"response_types"`
	GrantTypes                   []string         `tfsdk:"grant_types"`
	AllowedWebOrigins            []string         `tfsdk:"allowed_web_origins"`
	AllowedOrigins               []string         `tfsdk:"allowed_origins"`
	AllowedMfa                   []string         `tfsdk:"allowed_mfa"`
}

type RegistrationField struct {
	ID            types.String `tfsdk:"id"`
	Required      types.Bool   `tfsdk:"required"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	FieldKey      types.String `tfsdk:"field_key"`
	ConsentRefs   types.List   `tfsdk:"consent_refs"`
	DataType      types.String `tfsdk:"data_type"`
	ReadOnly      types.Bool   `tfsdk:"read_only"`
	Claimable     types.Bool   `tfsdk:"claimable"`
	ParentGroupId types.String `tfsdk:"parent_group_id"`
	Order         types.Int64  `tfsdk:"order"`
}

type TemplateGroup struct {
	ID                types.String `tfsdk:"id"`
	GroupId           types.String `tfsdk:"group_id"`
	SmsSenderConfig   types.Object `tfsdk:"sms_sender_config"`
	EmailSenderConfig types.Object `tfsdk:"email_sender_config"`
	IVRSenderConfig   types.Object `tfsdk:"ivr_sender_config"`
	PushSenderConfig  types.Object `tfsdk:"push_sender_config"`
}

type Template struct {
	ID             types.String `tfsdk:"id"`
	LastSeededBy   types.String `tfsdk:"last_seeded_by"`
	GroupId        types.String `tfsdk:"group_id"`
	TemplateKey    types.String `tfsdk:"template_key"`
	TemplateType   types.String `tfsdk:"template_type"`
	ProcessingType types.String `tfsdk:"processing_type"`
	Locale         types.String `tfsdk:"locale"`
	Language       types.String `tfsdk:"language"`
	UsageType      types.String `tfsdk:"usage_type"`
	Subject        types.String `tfsdk:"subject"`
	Content        types.String `tfsdk:"content"`
}
