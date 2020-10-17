// Copyright 2016 Circonus, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

// Key for CheckBundleConfig options and CheckDetails info
type Key string

// Constants per type as defined in
// https://login.circonus.com/resources/api/calls/check_bundle
const (
	//
	// default settings for api.NewCheckBundle()
	//
	DefaultCheckBundleMetricLimit = -1 // unlimited
	DefaultCheckBundleStatus      = "active"
	DefaultCheckBundlePeriod      = 60
	DefaultCheckBundleTimeout     = 10
	DefaultConfigOptionsSize      = 20

	//
	// common (apply to more than one check type)
	//
	AsyncMetrics       = Key("asynch_metrics")
	AuthMethod         = Key("auth_method")
	AuthPassword       = Key("auth_password")
	AuthUser           = Key("auth_user")
	BaseURL            = Key("base_url")
	CAChain            = Key("ca_chain")
	CertFile           = Key("certificate_file")
	Ciphers            = Key("ciphers")
	Command            = Key("command")
	DSN                = Key("dsn")
	HeaderPrefix       = Key("header_")
	HTTPVersion        = Key("http_version")
	KeyFile            = Key("key_file")
	Method             = Key("method")
	Password           = Key("password")
	Payload            = Key("payload")
	Port               = Key("port")
	Query              = Key("query")
	ReadLimit          = Key("read_limit")
	Secret             = Key("secret")
	SQL                = Key("sql")
	URI                = Key("uri")
	URL                = Key("url")
	Username           = Key("username")
	UseSSL             = Key("use_ssl")
	User               = Key("user")
	SASLAuthentication = Key("sasl_authentication")
	SASLUser           = Key("sasl_user")
	SecurityLevel      = Key("security_level")
	Version            = Key("version")
	AppendColumnName   = Key("append_column_name")
	Database           = Key("database")
	JDBCPrefix         = Key("jdbc_")

	//
	// CAQL check
	//
	// Common items:
	// Query

	//
	// Circonus Windows Agent
	//
	// Common items:
	// AuthPassword
	// AuthUser
	// Port
	// URL
	Calculated = Key("calculated")
	Category   = Key("category")

	//
	// Cloudwatch
	//
	// Notes:
	// DimPrefix is special because the actual key is dynamic and matches: `dim_(.+)`
	// Common items:
	// URL
	// Version
	APIKey            = Key("api_key")
	APISecret         = Key("api_secret")
	CloudwatchMetrics = Key("cloudwatch_metrics")
	DimPrefix         = Key("dim_")
	Granularity       = Key("granularity")
	Namespace         = Key("namespace")
	Statistics        = Key("statistics")

	//
	// Collectd
	//
	// Common items:
	// AsyncMetrics
	// Username
	// Secret
	// SecurityLevel

	//
	// Composite
	//
	CompositeMetricName = Key("composite_metric_name")
	Formula             = Key("formula")

	//
	// DHCP
	//
	HardwareAddress = Key("hardware_addr")
	HostIP          = Key("host_ip")
	RequestType     = Key("request_type")
	SendPort        = Key("send_port")

	//
	// DNS
	//
	// Common items:
	// Query
	CType      = Key("ctype")
	Nameserver = Key("nameserver")
	RType      = Key("rtype")

	//
	// EC Console
	//
	// Common items:
	// Command
	// Port
	// SASLAuthentication
	// SASLUser
	Objects = Key("objects")
	XPath   = Key("xpath")

	//
	// Elastic Search
	//
	// Common items:
	// Port
	// URL

	//
	// Ganglia
	//
	// Common items:
	// AsyncMetrics

	//
	// Google Analytics
	//
	// Common items:
	// Password
	// Username
	OAuthToken       = Key("oauth_token")
	OAuthTokenSecret = Key("oauth_token_secret")
	OAuthVersion     = Key("oauth_version")
	TableID          = Key("table_id")
	UseOAuth         = Key("use_oauth")

	//
	// HA Proxy
	//
	// Common items:
	// AuthPassword
	// AuthUser
	// Port
	// UseSSL
	Host   = Key("host")
	Select = Key("select")

	//
	// HTTP
	//
	// Notes:
	// HeaderPrefix is special because the actual key is dynamic and matches: `header_(\S+)`
	// Common items:
	// AuthMethod
	// AuthPassword
	// AuthUser
	// CAChain
	// CertFile
	// Ciphers
	// KeyFile
	// URL
	// HeaderPrefix
	// HTTPVersion
	// Method
	// Payload
	// ReadLimit
	Body      = Key("body")
	Code      = Key("code")
	Extract   = Key("extract")
	Redirects = Key("redirects")

	//
	// HTTPTRAP
	//
	// Common items:
	// AsyncMetrics
	// Secret

	//
	// IMAP
	//
	// Common items:
	// AuthPassword
	// AuthUser
	// CAChain
	// CertFile
	// Ciphers
	// KeyFile
	// Port
	// UseSSL
	Fetch      = Key("fetch")
	Folder     = Key("folder")
	HeaderHost = Key("header_Host")
	Search     = Key("search")

	//
	// JMX
	//
	// Common items:
	// Password
	// Port
	// URI
	// Username
	MbeanDomains = Key("mbean_domains")

	//
	// JSON
	//
	// Common items:
	// AuthMethod
	// AuthPassword
	// AuthUser
	// CAChain
	// CertFile
	// Ciphers
	// HeaderPrefix
	// HTTPVersion
	// KeyFile
	// Method
	// Payload
	// Port
	// ReadLimit
	// URL

	//
	// Keynote
	//
	// Notes:
	// SlotAliasPrefix is special because the actual key is dynamic and matches: `slot_alias_(\d+)`
	// Common items:
	// APIKey
	// BaseURL
	PageComponent   = Key("pagecomponent")
	SlotAliasPrefix = Key("slot_alias_")
	SlotIDList      = Key("slot_id_list")
	TransPageList   = Key("transpagelist")

	//
	// Keynote Pulse
	//
	// Common items:
	// BaseURL
	// Password
	// User
	AgreementID = Key("agreement_id")

	//
	// LDAP
	//
	// Common items:
	// Password
	// Port
	AuthType          = Key("authtype")
	DN                = Key("dn")
	SecurityPrincipal = Key("security_principal")

	//
	// Memcached
	//
	// Common items:
	// Port

	//
	// MongoDB
	//
	// Common items:
	// Command
	// Password
	// Port
	// Username
	DBName = Key("dbname")

	//
	// Munin
	//
	// Note: no configuration options

	//
	// MySQL
	//
	// Common items:
	// DSN
	// SQL

	//
	// Newrelic rpm
	//
	// Common items:
	// APIKey
	AccountID     = Key("acct_id")
	ApplicationID = Key("application_id")
	LicenseKey    = Key("license_key")

	//
	// Nginx
	//
	// Common items:
	// CAChain
	// CertFile
	// Ciphers
	// KeyFile
	// URL

	//
	// NRPE
	//
	// Common items:
	// Command
	// Port
	// UseSSL
	AppendUnits = Key("append_uom")

	//
	// NTP
	//
	// Common items:
	// Port
	Control = Key("control")

	//
	// Oracle
	//
	// Notes:
	// JDBCPrefix is special because the actual key is dynamic and matches: `jdbc_(\S+)`
	// Common items:
	// AppendColumnName
	// Database
	// JDBCPrefix
	// Password
	// Port
	// SQL
	// User

	//
	// Ping ICMP
	//
	AvailNeeded = Key("avail_needed")
	Count       = Key("count")
	Interval    = Key("interval")

	//
	// PostgreSQL
	//
	// Common items:
	// DSN
	// SQL

	//
	// Redis
	//
	// Common items:
	// Command
	// Password
	// Port
	DBIndex = Key("dbindex")

	//
	// Resmon
	//
	// Notes:
	// HeaderPrefix is special because the actual key is dynamic and matches: `header_(\S+)`
	// Common items:
	// AuthMethod
	// AuthPassword
	// AuthUser
	// CAChain
	// CertFile
	// Ciphers
	// HeaderPrefix
	// HTTPVersion
	// KeyFile
	// Method
	// Payload
	// Port
	// ReadLimit
	// URL

	//
	// SMTP
	//
	// Common items:
	// Payload
	// Port
	// SASLAuthentication
	// SASLUser
	EHLO               = Key("ehlo")
	From               = Key("from")
	ProxyDestAddress   = Key("proxy_dest_address")
	ProxyDestPort      = Key("proxy_dest_port")
	ProxyFamily        = Key("proxy_family")
	ProxyProtocol      = Key("proxy_protocol")
	ProxySourceAddress = Key("proxy_source_address")
	ProxySourcePort    = Key("proxy_source_port")
	SASLAuthID         = Key("sasl_auth_id")
	SASLPassword       = Key("sasl_password")
	StartTLS           = Key("starttls")
	To                 = Key("to")

	//
	// SNMP
	//
	// Notes:
	// OIDPrefix is special because the actual key is dynamic and matches: `oid_(.+)`
	// TypePrefix is special because the actual key is dynamic and matches: `type_(.+)`
	// Common items:
	// Port
	// SecurityLevel
	// Version
	AuthPassphrase    = Key("auth_passphrase")
	AuthProtocol      = Key("auth_protocol")
	Community         = Key("community")
	ContextEngine     = Key("context_engine")
	ContextName       = Key("context_name")
	OIDPrefix         = Key("oid_")
	PrivacyPassphrase = Key("privacy_passphrase")
	PrivacyProtocol   = Key("privacy_protocol")
	SecurityEngine    = Key("security_engine")
	SecurityName      = Key("security_name")
	SeparateQueries   = Key("separate_queries")
	TypePrefix        = Key("type_")

	//
	// SQLServer
	//
	// Notes:
	// JDBCPrefix is special because the actual key is dynamic and matches: `jdbc_(\S+)`
	// Common items:
	// AppendColumnName
	// Database
	// JDBCPrefix
	// Password
	// Port
	// SQL
	// User

	//
	// SSH v2
	//
	// Common items:
	// Port
	MethodCompCS      = Key("method_comp_cs")
	MethodCompSC      = Key("method_comp_sc")
	MethodCryptCS     = Key("method_crypt_cs")
	MethodCryptSC     = Key("method_crypt_sc")
	MethodHostKey     = Key("method_hostkey")
	MethodKeyExchange = Key("method_kex")
	MethodMacCS       = Key("method_mac_cs")
	MethodMacSC       = Key("method_mac_sc")

	//
	// StatsD
	//
	// Note: no configuration options

	//
	// TCP
	//
	// Common items:
	// CAChain
	// CertFile
	// Ciphers
	// KeyFile
	// Port
	// UseSSL
	BannerMatch = Key("banner_match")

	//
	// Varnish
	//
	// Note: no configuration options

	//
	// reserved - config option(s) can't actually be set - here for r/o access
	//
	ReverseSecretKey = Key("reverse:secret_key")
	SubmissionURL    = Key("submission_url")

	//
	// Endpoint prefix & cid regex
	//
	DefaultCIDRegex  = "[0-9]+"
	DefaultUUIDRegex = "[[:xdigit:]]{8}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{4}-[[:xdigit:]]{12}"

	// 2019-06-05: switching to an opaque cid validator, the cid format of rule_sets changed
	OpaqueCIDRegex             = ".+"
	AccountPrefix              = "/account"
	AccountCIDRegex            = "^(" + AccountPrefix + "/(" + OpaqueCIDRegex + "))$"
	AcknowledgementPrefix      = "/acknowledgement"
	AcknowledgementCIDRegex    = "^(" + AcknowledgementPrefix + "/(" + OpaqueCIDRegex + "))$"
	AlertPrefix                = "/alert"
	AlertCIDRegex              = "^(" + AlertPrefix + "/(" + OpaqueCIDRegex + "))$"
	AnnotationPrefix           = "/annotation"
	AnnotationCIDRegex         = "^(" + AnnotationPrefix + "/(" + OpaqueCIDRegex + "))$"
	BrokerPrefix               = "/broker"
	BrokerCIDRegex             = "^(" + BrokerPrefix + "/(" + OpaqueCIDRegex + "))$"
	CheckBundleMetricsPrefix   = "/check_bundle_metrics"
	CheckBundleMetricsCIDRegex = "^(" + CheckBundleMetricsPrefix + "/(" + OpaqueCIDRegex + "))$"
	CheckBundlePrefix          = "/check_bundle"
	CheckBundleCIDRegex        = "^(" + CheckBundlePrefix + "/(" + OpaqueCIDRegex + "))$"
	CheckPrefix                = "/check"
	CheckCIDRegex              = "^(" + CheckPrefix + "/(" + OpaqueCIDRegex + "))$"
	ContactGroupPrefix         = "/contact_group"
	ContactGroupCIDRegex       = "^(" + ContactGroupPrefix + "/(" + OpaqueCIDRegex + "))$"
	DashboardPrefix            = "/dashboard"
	DashboardCIDRegex          = "^(" + DashboardPrefix + "/(" + OpaqueCIDRegex + "))$"
	GraphPrefix                = "/graph"
	GraphCIDRegex              = "^(" + GraphPrefix + "/(" + OpaqueCIDRegex + "))$"
	MaintenancePrefix          = "/maintenance"
	MaintenanceCIDRegex        = "^(" + MaintenancePrefix + "/(" + OpaqueCIDRegex + "))$"
	MetricClusterPrefix        = "/metric_cluster"
	MetricClusterCIDRegex      = "^(" + MetricClusterPrefix + "/(" + OpaqueCIDRegex + "))$"
	MetricPrefix               = "/metric"
	MetricCIDRegex             = "^(" + MetricPrefix + "/(" + OpaqueCIDRegex + "))$"
	OutlierReportPrefix        = "/outlier_report"
	OutlierReportCIDRegex      = "^(" + OutlierReportPrefix + "/(" + OpaqueCIDRegex + "))$"
	ProvisionBrokerPrefix      = "/provision_broker"
	ProvisionBrokerCIDRegex    = "^(" + ProvisionBrokerPrefix + "/(" + OpaqueCIDRegex + "))$"
	RuleSetGroupPrefix         = "/rule_set_group"
	RuleSetGroupCIDRegex       = "^(" + RuleSetGroupPrefix + "/(" + OpaqueCIDRegex + "))$"
	RuleSetPrefix              = "/rule_set"
	RuleSetCIDRegex            = "^(" + RuleSetPrefix + "/(" + OpaqueCIDRegex + "))$"
	UserPrefix                 = "/user"
	UserCIDRegex               = "^(" + UserPrefix + "/(" + OpaqueCIDRegex + "))$"
	WorksheetPrefix            = "/worksheet"
	WorksheetCIDRegex          = "^(" + WorksheetPrefix + "/(" + OpaqueCIDRegex + "))$"

	// contact group serverity levels
	NumSeverityLevels = 5

	// AccountCIDRegex            = "^(" + AccountPrefix + "/(" + DefaultCIDRegex + "|current))$"
	// AcknowledgementCIDRegex    = "^(" + AcknowledgementPrefix + "/(" + DefaultCIDRegex + "))$"
	// AlertCIDRegex              = "^(" + AlertPrefix + "/(" + DefaultCIDRegex + "))$"
	// AnnotationCIDRegex         = "^(" + AnnotationPrefix + "/(" + DefaultCIDRegex + "))$"
	// BrokerCIDRegex             = "^(" + BrokerPrefix + "/(" + DefaultCIDRegex + "))$"
	// CheckBundleMetricsCIDRegex = "^(" + CheckBundleMetricsPrefix + "/(" + DefaultCIDRegex + "))$"
	// CheckBundleCIDRegex        = "^(" + CheckBundlePrefix + "/(" + DefaultCIDRegex + "))$"
	// CheckCIDRegex              = "^(" + CheckPrefix + "/(" + DefaultCIDRegex + "))$"
	// ContactGroupCIDRegex       = "^(" + ContactGroupPrefix + "/(" + DefaultCIDRegex + "))$"
	// DashboardCIDRegex          = "^(" + DashboardPrefix + "/(" + DefaultCIDRegex + "))$"
	// GraphCIDRegex              = "^(" + GraphPrefix + "/(" + DefaultUUIDRegex + "))$"
	// MaintenanceCIDRegex        = "^(" + MaintenancePrefix + "/(" + DefaultCIDRegex + "))$"
	// MetricClusterCIDRegex      = "^(" + MetricClusterPrefix + "/(" + DefaultCIDRegex + "))$"
	// MetricCIDRegex             = "^(" + MetricPrefix + "/((" + DefaultCIDRegex + ")_([^[:space:]]+)))$"
	// OutlierReportCIDRegex      = "^(" + OutlierReportPrefix + "/(" + DefaultCIDRegex + "))$"
	// ProvisionBrokerCIDRegex    = "^(" + ProvisionBrokerPrefix + "/([a-z0-9]+-[a-z0-9]+))$"
	// RuleSetGroupCIDRegex       = "^(" + RuleSetGroupPrefix + "/(" + DefaultCIDRegex + "))$"
	// RuleSetCIDRegex            = "^(" + RuleSetPrefix + "/((" + DefaultCIDRegex + ")_([^[:space:]]+)))$"
	// UserCIDRegex               = "^(" + UserPrefix + "/(" + DefaultCIDRegex + "|current))$"
	// WorksheetCIDRegex          = "^(" + WorksheetPrefix + "/(" + DefaultUUIDRegex + "))$"
)
