# Thunderbird Auto Configuration Server (TACS)
Provides a simple Thunderbird configuration web server with templating by group or login

## CLI
The compiled application has built-in help and configuration templates:
```bash
./tacs -h
Thunderbird AutoConfig Server (TACS) help.
Environment variables:
  TACS_SCHEME                             - specifies a path to the definition scheme
  TACS_ADDR                               - dns or IP address of the http server. You can leave it unspecified,
                                          then listening will be performed on all available interfaces
  TACS_PORT                               - listening http port
  TACS_CERT и TACS_CERT_KEY               - specifies the path to the SSL certificate and SSL key, respectively. 
                                          If both are specified (not equal to ""), the server will try to start using SSL
  TACS_LDAP_SERVER                        - ldap server host
  TACS_LDAP_PORT                          - ldap server port
  TACS_LDAP_CERT и TACS_LDAP_KEY          - point to the SSL certificate and SSL key, respectively. 
                                          If both are specified (not equal to ""), an attempt will be made to switch LDAP to SSL
  TACS_LDAP_USER и TACS_LDAP_PASSWORD     - specify user credentials with LDAP read access

Program arguments:
  -example-scheme
        print example scheme config
  -help
        shows startup help (this)
```
Getting a configuration template:
```bash
./tacs -example-scheme > scheme.yaml
cat ./scheme.yaml
---
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
  #searchBase: OU=SystemUsers,OU=Corp,DC=corp,DC=domain,DC=com

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
    #    photo: raw:jpegPhoto
```

Starting the server:
```bash
tree ./
.
├── scheme.yaml
├── tacs
└── templates
    ├── base.go.tmpl
    ├── default.go.tmpl
    ├── managers.tmpl
    └── ...etс...
```
```bash
TACS_SCHEME=scheme.yaml \
TACS_PORT=8080 \
TACS_LDAP_SERVER=ldap.example.com \
TACS_LDAP_PORT=389 \
TACS_LDAP_USER=ldap@corp.example.com \
TACS_LDAP_PASSWORD=Qwerty \
./tacs
```

## Templates
### General
The directory for searching for templates is specified in `scheme.yaml`
```YAML
---
templateDir: templates
```
TACS traverses all subdirectories except symbolic links, analyzing and loading *.tmpl files into memory.

To get acquainted with go-templates, read [documentation](https://pkg.go.dev/text/template).

Below are the features of working with templates in TACS:
- Template declaration is done by the `{{define "template_name"}}...{{end}}` block.
- The template name is specified in `scheme.yaml`, in the `template` property and is used when generating the configuration page:
  ```YAML
  local:
    default:
      template: template_name
  ...
    list:
      username:
        template: template_name
  ...
  ldap:
    default:
      template: template_name
  ...
    list:
      - group: ...
        template: template_name
  ```
- Templates can call other templates:
  ```yaml
  {{define "temp1"}}
  ...
  {{end}}
  {{define "temp2"}}
  ...
  {{end}}
  {{define "temp3"}}
  {{template "temp1"}}
  {{template "temp2"}}
  {{end}}
  ```
- The template name must be unique compared to others, otherwise templates may overwrite each other.
- You can use variables in the template to substitute values from `scheme.yaml`/LDAP:
  ```yaml
  # All variables are stored at "dot":
  {{define "default"}}
  {{.var_name}}
  {{end}}
  ```
- Variables are declared in `scheme.yaml`, in the `fields` property:
  ```yaml
  local:
    default:
      template_key: value
  ...
  list:
    username:
      template_key: value
  ...
  ldap:
  default:
    fields:
      template_key: ldap_field_with_value
      template_key: raw:ldap_field_with_binary_value
  ...
  list:
    - group: ...
      fields:
        template_key: ldap_field_with_value
  ```
### Templates + configuration for Thunderbird
Using [PrefApi]((#prefapi)), prepares the [template](#general) (example `default.tmpl`)

The basis for it can be the `pref.js` file in the directory of the **already configured** thunderbird profile:
- Windows - `%APPDATA%\Thunderbird\Profiles\*\prefs.js`
- Linux: - `~/.thunderbird/*/prefs.js`

> **Attention!** 
> Changing the `pref.js` file itself is useless, since it is updated by the mail client at runtime.

# Thunderbird
Below there will be **brief** excerpts from various sources that will allow you to prepare your email client for automatic configuration.

Sources:
- [The SIPB Thunderbird Locker - Maintainers
](https://web.mit.edu/~thunderbird/www/maintainers/autoconfig.html)
- [MCD, Mission Control Desktop, AKA AutoConfig](https://udn.realityripple.com/docs/Archive/Misc_top_level/MCD,_Mission_Control_Desktop_AKA_AutoConfig)
- [EASY THUNDERBIRD ACCOUNT MANAGEMENT USING MCD](https://blog.deanandadie.net/2010/06/easy-thunderbird-account-management-using-mcd/)
- [Идеальный корпоративный почтовый клиент
](https://habr.com/ru/articles/101905/)
- [Customizing Firefox Using AutoConfig](https://support.mozilla.org/en-US/kb/customizing-firefox-using-autoconfig)

## PrefApi
are configuration functions that control the parameters of the mail client when it starts.

The PrefApi functions are defined in the file `$THUNDERBIRD_FOLDER/omni.ja/defaults/autoconfig/prefcalls.js`:
- `getPrefBranch()` - gets the root of the preference tree. **Not directly used**.
- `pref(prefName, value)` - changes the current value to the specified one. **Most often used**.
- `defaultPref(prefName, value)` - sets the parameter to its default value.
   That is if the user leaves the field blank or resets the settings, the specified value will be set.
- `lockPref(prefName, value)` - locks a parameter at the specified value.
    Users cannot change locked parameters.
- `unlockPref(prefName)` - разблокирует, ранее была заблокированный, параметр.
- `getPref(prefName)` - получает значение указанного параметра.
- `displayError(funcname, message)`- displays an error dialog when starting Thunderbird. **The only means of debugging variables**.
- `getenv(name)` - gets the value of the specified environment variable from the user's environment.
- Functions for working with LDAP. They don't support authorization, so they're useless:
  - `setLDAPVersion(version)` - sets the LDAP version used by the server.
  - `getLDAPAttributes(host, base, filter, attribs)` - retrieves LDAP attributes from a given server.
  - `getLDAPValue(str, key)` - gets the LDAP value for a given row, filtered by key.

## Configuration files
### thunderbird.cfg
The settings file is JavaScript, so it has access to variables, functions, etc.
To use an external source of settings, write the `thunderbird.cfg` script:
```js
// Loading settings from the server
try {
    // Getting username from OS
    if (getenv("USER") != "") {
        // *NIX settings
        var env_user = getenv("USER");
    } else {
        // Windows settings
        var env_user = getenv("USERNAME");
    }
// Specifying a configuration server
pref("autoadmin.global_config_url", "http://server_host/"+env_user);
// Don't add mail variable to request
pref("autoadmin.append_emailaddr", false);
 
} catch (e) {
    displayError("pref", e);
}
```
 and placed in the program directory:
- Windows - `C:\Program Files (x86)\Mozilla Thunderbird\thunderbird.cfg`
- Linux - `/usr/lib/thunderbird/thunderbird.cfg`

### autoconfig.js
In order for `thunderbird.cfg` to be used, you need to create a file `autoconfig.js`:
```js
// Specify the configuration file
pref('general.config.filename', 'thunderbird.cfg');
// Disabling bit shifting to read a regular file
pref('general.config.obscure_value', 0);
```
and put it on the way:
- Windows - `C:\Program Files\Mozilla Thunderbird\defaults\pref\autoconf.js`
- Linux - `/usr/share/thunderbird/defaults/pref/autoconfig.js`


# Security
The code was checked by static analyzers that are focused on finding vulnerabilities.

## GoSec
Installation:
```sh
go install github.com/securego/gosec/v2/cmd/gosec@latest
```
Launch:
```sh
gosec ./...
```
## Go Vulnerability
Installation:
```sh
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Launch:
```sh
govulncheck ./...
```
