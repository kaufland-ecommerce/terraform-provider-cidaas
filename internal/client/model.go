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

type HostedPageGroup struct {
	Name  string
	Pages map[string]string
}

type AppKey struct {
	ID         string `json:"id"`
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type App struct {
	ID                               string  `json:"id"`
	ClientId                         string  `json:"client_id"`
	ClientSecret                     string  `json:"client_secret"`
	ClientName                       string  `json:"client_name"`
	ClientDisplayName                string  `json:"client_display_name"`
	IsRememberMeSelected             bool    `json:"is_remember_me_selected"`
	ClientType                       string  `json:"client_type"`
	AllowDisposableEmail             bool    `json:"allow_disposable_email"`
	FdsEnabled                       bool    `json:"fds_enabled"`
	EnablePasswordlessAuth           bool    `json:"enable_passwordless_auth"`
	EnableDeduplication              bool    `json:"enable_deduplication"`
	MobileNumberVerificationRequired bool    `json:"mobile_number_verification_required"`
	HostedPageGroup                  string  `json:"hosted_page_group"`
	PrimaryColor                     string  `json:"primaryColor"`
	AccentColor                      string  `json:"accentColor"`
	AutoLoginAfterRegister           bool    `json:"auto_login_after_register"`
	CompanyName                      string  `json:"company_name"`
	CompanyAddress                   string  `json:"company_address"`
	CompanyWebsite                   string  `json:"company_website"`
	TokenLifetimeInSeconds           int64   `json:"token_lifetime_in_seconds"`
	IdTokenLifetimeInSeconds         int64   `json:"id_token_lifetime_in_seconds"`
	RefreshTokenLifetimeInSeconds    int64   `json:"refresh_token_lifetime_in_seconds"`
	EmailVerificationRequired        bool    `json:"email_verification_required"`
	EnableBotDetection               bool    `json:"enable_bot_detection"`
	IsLoginSuccessPageEnabled        bool    `json:"is_login_success_page_enabled"`
	AllowGuestLogin                  bool    `json:"allow_guest_login"`
	JweEnabled                       bool    `json:"jwe_enabled"`
	AlwaysAskMfa                     bool    `json:"always_ask_mfa"`
	PasswordPolicy                   *string `json:"password_policy_ref,omitempty"`

	AppKey AppKey `json:"appKey"`

	AllowLoginWith               []string         `json:"allow_login_with"`
	RedirectUris                 []string         `json:"redirect_uris"`
	AllowedLogoutUrls            []string         `json:"allowed_logout_urls"`
	AllowedScopes                []string         `json:"allowed_scopes"`
	SocialProviders              []SocialProvider `json:"social_providers"`
	AdditionalAccessTokenPayload []string         `json:"additional_access_token_payload"`
	AllowedFields                []string         `json:"allowed_fields"`
	RequiredFields               []string         `json:"required_fields"`
	ConsentRefs                  []string         `json:"consent_refs"`
	ResponseTypes                []string         `json:"response_types"`
	GrantTypes                   []string         `json:"grant_types"`
	AllowedWebOrigins            []string         `json:"allowed_web_origins"`
	AllowedOrigins               []string         `json:"allowed_origins"`
	AllowedMfa                   []string         `json:"allowed_mfa"`
}
