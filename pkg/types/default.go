package types

const (
	DefaultHTTPMessage = "The collaborator server received an {{protocol}} request from {{from}} at {{time}}:\n```\n{{request}}\n{{response}}```"
	DefaultDNSMessage  = "The collaborator server received a DNS lookup of type {{type}} for the domain name {{domain}} from {{from}} at {{time}}:\n```{{request}}```"
	DefaultSMTPMessage = "The collaborator server received an SMTP connection from IP address {{from}} at {{time}}\n\nThe email details were:\n\nFrom:\n{{sender}}\n\nTo:\n{{recipients}}\n\nMessage:\n{{message}}\n\nSMTP Conversation:\n{{conversation}}"
	DefaultCLIMessage  = "{{data}}"
)
