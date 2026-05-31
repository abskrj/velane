package providers

// CatalogEntry is lightweight provider metadata used for the Integrations
// listing page — name, slug, auth mode, and category tags.
// Rich per-provider docs (endpoints, examples) live in providers_meta.json.
type CatalogEntry struct {
	UniqueKey   string   `json:"unique_key"`
	Name        string   `json:"name"`
	AuthMode    string   `json:"auth_mode"`
	Categories  []string `json:"categories"`
}

// Catalog is the full list of available providers returned by GET /v1/integrations.
// Populated from the bundled static list below — no runtime Nango call needed.
var Catalog = []CatalogEntry{
	// CRM
	{UniqueKey: "salesforce", Name: "Salesforce", AuthMode: "OAUTH2", Categories: []string{"crm"}},
	{UniqueKey: "hubspot", Name: "HubSpot", AuthMode: "OAUTH2", Categories: []string{"crm"}},
	{UniqueKey: "pipedrive", Name: "Pipedrive", AuthMode: "OAUTH2", Categories: []string{"crm"}},
	{UniqueKey: "zoho-crm", Name: "Zoho CRM", AuthMode: "OAUTH2", Categories: []string{"crm"}},
	{UniqueKey: "close", Name: "Close", AuthMode: "API_KEY", Categories: []string{"crm"}},
	{UniqueKey: "copper", Name: "Copper", AuthMode: "API_KEY", Categories: []string{"crm"}},
	{UniqueKey: "freshsales", Name: "Freshsales", AuthMode: "API_KEY", Categories: []string{"crm"}},
	{UniqueKey: "affinity", Name: "Affinity", AuthMode: "API_KEY", Categories: []string{"crm"}},
	// Developer Tools
	{UniqueKey: "github", Name: "GitHub", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "gitlab", Name: "GitLab", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "bitbucket", Name: "Bitbucket", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "jira", Name: "Jira", AuthMode: "OAUTH2", Categories: []string{"developer-tools", "project-management"}},
	{UniqueKey: "linear", Name: "Linear", AuthMode: "OAUTH2", Categories: []string{"developer-tools", "project-management"}},
	{UniqueKey: "asana", Name: "Asana", AuthMode: "OAUTH2", Categories: []string{"developer-tools", "project-management"}},
	{UniqueKey: "clickup", Name: "ClickUp", AuthMode: "OAUTH2", Categories: []string{"developer-tools", "project-management"}},
	{UniqueKey: "shortcut", Name: "Shortcut", AuthMode: "API_KEY", Categories: []string{"developer-tools"}},
	{UniqueKey: "pagerduty", Name: "PagerDuty", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "datadog", Name: "Datadog", AuthMode: "API_KEY", Categories: []string{"developer-tools"}},
	{UniqueKey: "sentry", Name: "Sentry", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "vercel", Name: "Vercel", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "render", Name: "Render", AuthMode: "API_KEY", Categories: []string{"developer-tools"}},
	// Communication
	{UniqueKey: "slack", Name: "Slack", AuthMode: "OAUTH2", Categories: []string{"communication"}},
	{UniqueKey: "discord", Name: "Discord", AuthMode: "OAUTH2", Categories: []string{"communication"}},
	{UniqueKey: "microsoft-teams", Name: "Microsoft Teams", AuthMode: "OAUTH2", Categories: []string{"communication"}},
	{UniqueKey: "intercom", Name: "Intercom", AuthMode: "OAUTH2", Categories: []string{"communication", "support"}},
	{UniqueKey: "twilio", Name: "Twilio", AuthMode: "API_KEY", Categories: []string{"communication"}},
	{UniqueKey: "sendgrid", Name: "SendGrid", AuthMode: "API_KEY", Categories: []string{"communication", "email"}},
	{UniqueKey: "mailchimp", Name: "Mailchimp", AuthMode: "OAUTH2", Categories: []string{"communication", "email", "marketing"}},
	{UniqueKey: "mailgun", Name: "Mailgun", AuthMode: "API_KEY", Categories: []string{"communication", "email"}},
	{UniqueKey: "postmark", Name: "Postmark", AuthMode: "API_KEY", Categories: []string{"communication", "email"}},
	{UniqueKey: "zoom", Name: "Zoom", AuthMode: "OAUTH2", Categories: []string{"communication"}},
	{UniqueKey: "webex", Name: "Webex", AuthMode: "OAUTH2", Categories: []string{"communication"}},
	// Email / Calendar
	{UniqueKey: "gmail", Name: "Gmail", AuthMode: "OAUTH2", Categories: []string{"email"}},
	{UniqueKey: "google-calendar", Name: "Google Calendar", AuthMode: "OAUTH2", Categories: []string{"calendar"}},
	{UniqueKey: "outlook", Name: "Outlook", AuthMode: "OAUTH2", Categories: []string{"email", "calendar"}},
	{UniqueKey: "microsoft-calendar", Name: "Microsoft Calendar", AuthMode: "OAUTH2", Categories: []string{"calendar"}},
	{UniqueKey: "calendly", Name: "Calendly", AuthMode: "OAUTH2", Categories: []string{"calendar"}},
	// Storage / Documents
	{UniqueKey: "google-drive", Name: "Google Drive", AuthMode: "OAUTH2", Categories: []string{"storage"}},
	{UniqueKey: "google-sheets", Name: "Google Sheets", AuthMode: "OAUTH2", Categories: []string{"storage", "spreadsheets"}},
	{UniqueKey: "google-docs", Name: "Google Docs", AuthMode: "OAUTH2", Categories: []string{"storage", "documents"}},
	{UniqueKey: "dropbox", Name: "Dropbox", AuthMode: "OAUTH2", Categories: []string{"storage"}},
	{UniqueKey: "box", Name: "Box", AuthMode: "OAUTH2", Categories: []string{"storage"}},
	{UniqueKey: "onedrive", Name: "OneDrive", AuthMode: "OAUTH2", Categories: []string{"storage"}},
	{UniqueKey: "sharepoint", Name: "SharePoint", AuthMode: "OAUTH2", Categories: []string{"storage", "documents"}},
	{UniqueKey: "notion", Name: "Notion", AuthMode: "OAUTH2", Categories: []string{"productivity", "documents"}},
	{UniqueKey: "confluence", Name: "Confluence", AuthMode: "OAUTH2", Categories: []string{"documents"}},
	// Project Management
	{UniqueKey: "trello", Name: "Trello", AuthMode: "OAUTH2", Categories: []string{"project-management"}},
	{UniqueKey: "monday", Name: "Monday.com", AuthMode: "OAUTH2", Categories: []string{"project-management"}},
	{UniqueKey: "wrike", Name: "Wrike", AuthMode: "OAUTH2", Categories: []string{"project-management"}},
	{UniqueKey: "basecamp", Name: "Basecamp", AuthMode: "OAUTH2", Categories: []string{"project-management"}},
	{UniqueKey: "smartsheet", Name: "Smartsheet", AuthMode: "OAUTH2", Categories: []string{"project-management"}},
	{UniqueKey: "airtable", Name: "Airtable", AuthMode: "OAUTH2", Categories: []string{"productivity", "spreadsheets"}},
	{UniqueKey: "todoist", Name: "Todoist", AuthMode: "OAUTH2", Categories: []string{"productivity"}},
	// Payments / Finance
	{UniqueKey: "stripe", Name: "Stripe", AuthMode: "API_KEY", Categories: []string{"payments"}},
	{UniqueKey: "shopify", Name: "Shopify", AuthMode: "OAUTH2", Categories: []string{"payments", "ecommerce"}},
	{UniqueKey: "paypal", Name: "PayPal", AuthMode: "OAUTH2", Categories: []string{"payments"}},
	{UniqueKey: "chargebee", Name: "Chargebee", AuthMode: "API_KEY", Categories: []string{"payments"}},
	{UniqueKey: "recurly", Name: "Recurly", AuthMode: "API_KEY", Categories: []string{"payments"}},
	{UniqueKey: "braintree", Name: "Braintree", AuthMode: "API_KEY", Categories: []string{"payments"}},
	{UniqueKey: "quickbooks", Name: "QuickBooks", AuthMode: "OAUTH2", Categories: []string{"finance"}},
	{UniqueKey: "xero", Name: "Xero", AuthMode: "OAUTH2", Categories: []string{"finance"}},
	{UniqueKey: "ramp", Name: "Ramp", AuthMode: "OAUTH2", Categories: []string{"finance"}},
	// Support
	{UniqueKey: "zendesk", Name: "Zendesk", AuthMode: "OAUTH2", Categories: []string{"support"}},
	{UniqueKey: "freshdesk", Name: "Freshdesk", AuthMode: "API_KEY", Categories: []string{"support"}},
	{UniqueKey: "helpscout", Name: "Help Scout", AuthMode: "OAUTH2", Categories: []string{"support"}},
	{UniqueKey: "front", Name: "Front", AuthMode: "OAUTH2", Categories: []string{"support", "communication"}},
	{UniqueKey: "re-amaze", Name: "Re:amaze", AuthMode: "API_KEY", Categories: []string{"support"}},
	// Marketing
	{UniqueKey: "marketo", Name: "Marketo", AuthMode: "OAUTH2", Categories: []string{"marketing"}},
	{UniqueKey: "braze", Name: "Braze", AuthMode: "API_KEY", Categories: []string{"marketing"}},
	{UniqueKey: "klaviyo", Name: "Klaviyo", AuthMode: "API_KEY", Categories: []string{"marketing", "email"}},
	{UniqueKey: "segment", Name: "Segment", AuthMode: "API_KEY", Categories: []string{"marketing", "analytics"}},
	{UniqueKey: "mixpanel", Name: "Mixpanel", AuthMode: "API_KEY", Categories: []string{"analytics"}},
	{UniqueKey: "amplitude", Name: "Amplitude", AuthMode: "API_KEY", Categories: []string{"analytics"}},
	{UniqueKey: "customer-io", Name: "Customer.io", AuthMode: "API_KEY", Categories: []string{"marketing", "email"}},
	{UniqueKey: "activecampaign", Name: "ActiveCampaign", AuthMode: "API_KEY", Categories: []string{"marketing", "email"}},
	{UniqueKey: "convertkit", Name: "ConvertKit", AuthMode: "API_KEY", Categories: []string{"marketing", "email"}},
	{UniqueKey: "pardot", Name: "Pardot", AuthMode: "OAUTH2", Categories: []string{"marketing"}},
	// HR
	{UniqueKey: "workday", Name: "Workday", AuthMode: "OAUTH2", Categories: []string{"hr"}},
	{UniqueKey: "bamboohr", Name: "BambooHR", AuthMode: "API_KEY", Categories: []string{"hr"}},
	{UniqueKey: "gusto", Name: "Gusto", AuthMode: "OAUTH2", Categories: []string{"hr"}},
	{UniqueKey: "rippling", Name: "Rippling", AuthMode: "OAUTH2", Categories: []string{"hr"}},
	{UniqueKey: "greenhouse", Name: "Greenhouse", AuthMode: "API_KEY", Categories: []string{"hr", "recruiting"}},
	{UniqueKey: "lever", Name: "Lever", AuthMode: "OAUTH2", Categories: []string{"hr", "recruiting"}},
	{UniqueKey: "merge", Name: "Merge", AuthMode: "API_KEY", Categories: []string{"hr"}},
	// Data / Analytics
	{UniqueKey: "snowflake", Name: "Snowflake", AuthMode: "API_KEY", Categories: []string{"data"}},
	{UniqueKey: "bigquery", Name: "BigQuery", AuthMode: "OAUTH2", Categories: []string{"data"}},
	{UniqueKey: "redshift", Name: "Redshift", AuthMode: "API_KEY", Categories: []string{"data"}},
	{UniqueKey: "databricks", Name: "Databricks", AuthMode: "API_KEY", Categories: []string{"data"}},
	{UniqueKey: "fivetran", Name: "Fivetran", AuthMode: "API_KEY", Categories: []string{"data"}},
	// Security / Identity
	{UniqueKey: "okta", Name: "Okta", AuthMode: "OAUTH2", Categories: []string{"security", "identity"}},
	{UniqueKey: "auth0", Name: "Auth0", AuthMode: "OAUTH2", Categories: []string{"security", "identity"}},
	{UniqueKey: "onelogin", Name: "OneLogin", AuthMode: "OAUTH2", Categories: []string{"security", "identity"}},
	// Social
	{UniqueKey: "twitter", Name: "Twitter / X", AuthMode: "OAUTH2", Categories: []string{"social"}},
	{UniqueKey: "linkedin", Name: "LinkedIn", AuthMode: "OAUTH2", Categories: []string{"social"}},
	{UniqueKey: "facebook", Name: "Facebook", AuthMode: "OAUTH2", Categories: []string{"social"}},
	{UniqueKey: "instagram", Name: "Instagram", AuthMode: "OAUTH2", Categories: []string{"social"}},
	// Logistics / E-commerce
	{UniqueKey: "woocommerce", Name: "WooCommerce", AuthMode: "API_KEY", Categories: []string{"ecommerce"}},
	{UniqueKey: "bigcommerce", Name: "BigCommerce", AuthMode: "OAUTH2", Categories: []string{"ecommerce"}},
	{UniqueKey: "magento", Name: "Magento", AuthMode: "OAUTH2", Categories: []string{"ecommerce"}},
	// Other popular
	{UniqueKey: "figma", Name: "Figma", AuthMode: "OAUTH2", Categories: []string{"design"}},
	{UniqueKey: "canva", Name: "Canva", AuthMode: "OAUTH2", Categories: []string{"design"}},
	{UniqueKey: "miro", Name: "Miro", AuthMode: "OAUTH2", Categories: []string{"productivity"}},
	{UniqueKey: "loom", Name: "Loom", AuthMode: "OAUTH2", Categories: []string{"productivity"}},
	{UniqueKey: "typeform", Name: "Typeform", AuthMode: "OAUTH2", Categories: []string{"productivity"}},
	{UniqueKey: "surveymonkey", Name: "SurveyMonkey", AuthMode: "OAUTH2", Categories: []string{"productivity"}},
	{UniqueKey: "webflow", Name: "Webflow", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "wordpress", Name: "WordPress", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "docusign", Name: "DocuSign", AuthMode: "OAUTH2", Categories: []string{"documents"}},
	{UniqueKey: "hellosign", Name: "HelloSign", AuthMode: "OAUTH2", Categories: []string{"documents"}},
	{UniqueKey: "algolia", Name: "Algolia", AuthMode: "API_KEY", Categories: []string{"developer-tools"}},
	{UniqueKey: "contentful", Name: "Contentful", AuthMode: "OAUTH2", Categories: []string{"developer-tools"}},
	{UniqueKey: "sanity", Name: "Sanity", AuthMode: "API_KEY", Categories: []string{"developer-tools"}},
	{UniqueKey: "retool", Name: "Retool", AuthMode: "API_KEY", Categories: []string{"developer-tools"}},
	{UniqueKey: "supabase", Name: "Supabase", AuthMode: "API_KEY", Categories: []string{"developer-tools"}},
}
