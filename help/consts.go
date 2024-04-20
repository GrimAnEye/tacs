package help

const exampleScheme = `---
# Directory to search for templates
# The search is also performed in subdirectories, but without following symbolic links
# Required file extension: *.tmpl
#templateDir: templates

# Block for local processing of user logins.
# Executes first because it doesn't require external requests.
# If there is a match, subsequent stages are not checked
#local:

  # Default fields and properties
  #default:

  # Name of the GO template (*.tmpl), with insert fields - {{define "templateName"}}
  #template: default

  # A key-value map (hash table) that is used to fill the template.
  # Can be overridden at the user level.
  # If the field is not present at the user level, it will be set to the specified value.
  # The key is a variable declared in the template - {{ .variableName }}
  # Value - what will be substituted for the key
  #fields:
  #  cn: user
  #  mail: mail@example.com
  #  telephoneNumber:
  #  position: no_body
  #  company: "\"Example\""

  # List of users to handle requests.
  # Is a hash table of "username: params",
  # so each iteration of the key overwrites the previous values
  #list:
  #root:
  #  template: root
  #  fields:
  #    company: ""
  #    mail: root@example.com
  #    cn: Administrator
  #    signatureIsHTML: true
  #    signature: 'Sincerely\nroot​​​​​​'

# Block for processing users from LDAP.
# If there is a match, subsequent stages are not checked
#ldap:

  # A unique field that identifies the user, a filter compiled for this:
  #  (uid=username), example - (sAMAccountName=username)
  #uid: sAMAccountName

  # Ldap path where to start the search
  #usersSearchBase: OU=SystemUsers,OU=Corp,DC=corp,DC=domain,DC=com
  #groupsSearchBase: OU=SystemUsers,OU=Corp,DC=corp,DC=domain,DC=com

  # Additional filter for user search. Added to the filter with uid:
  # (&(uid=username)filter)
  # In this case, search only for enabled users:
  #filter: (!(userAccountControl:1.2.840.113556.1.4.803:=2))(objectCategory=person)(objectClass=user)

  # Search in subgroups?
  # If true - adds the option ":1.2.840.113556.1.4.1941:" to the filter
  #subgroups: true

  # Should I generate a response for a user who is not a member of any declared group?
  # In this case, a default section must be declared
  #allowWithoutGroups: true

  # A key-value map (hash table) that is used to fill the template.
  # Can be overridden at the ldap group level.
  # If the field is not present at the user level, it will be set to the specified value.
  # The key is a variable declared in the template - {{ .variableName }}
  # Value - fields taken from the user's LDAP profile.
  # Note.
  # If you prefix an LDAP field name with "raw:", the value will be queried raw and base64 encoded
  #default:
  #  template: default
  #  fields:
  #    cn: cn
  #    mail: mail
  #    telephoneNumber: telephoneNumber
  #    position: position
  #    company: company
  #    photo: raw:jpegPhoto
  #list:
    # The list of groups by which the user's groups are matched.
    # The check is performed one by one, and the first match
    # interrupts further processing.
    #- group: "CN=Domain Users,OU=SystemUsers,OU=Corp,DC=corp,DC=domain,DC=com"
    #  template: default
    #  fields:
    #    cn: cn
    #    mail: mail
    #    telephoneNumber: telephoneNumber
    #    position: position
    #    company: company
    #    photo: raw:jpegPhoto`

const help = `Thunderbird AutoConfig Server (TACS) help.
Environment variables:
	TACS_SCHEME				- specifies a path to the definition scheme
	TACS_ADDR				- dns or IP address of the http server. You can leave it unspecified,
						then listening will be performed on all available interfaces
	TACS_PORT				- listening http port
	TACS_CERT и TACS_CERT_KEY		- specifies the path to the SSL certificate and SSL key, respectively. 
						If both are specified (not equal to ""), the server will try to start using SSL
	TACS_LDAP_SERVER			- ldap server host
	TACS_LDAP_PORT				- ldap server port
	TACS_LDAP_CERT и TACS_LDAP_KEY		- point to the SSL certificate and SSL key, respectively. 
						If both are specified (not equal to ""), an attempt will be made to switch LDAP to SSL
	TACS_LDAP_USER и TACS_LDAP_PASSWORD	- specify user credentials with LDAP read access

Program arguments:`
