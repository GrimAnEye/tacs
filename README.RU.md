# Thunderbird Auto Configuration Server (TACS)
Предоставляет простой web-сервер конфигурации Thunderbird с шаблонизацией по группам или логину

## Командная строка
Скомпилированное приложение имеет встроенную справку и шаблоны конфигурации:
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
Получение шаблонна конфигурации:
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

Запуск сервера:
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

## Шаблоны
### Общее
Каталог для поиска шаблонов указывается в `scheme.yaml`
```YAML
---
templateDir: templates
```
TACS проходит по всем подкаталогам, кроме символических ссылок, анализируя и загружая в память `*.tmpl` файлы.

Для знакомства с go-шаблонами, читай [документацию](https://pkg.go.dev/text/template).

Ниже представлены особенности работы с шаблонами в TACS:
- Объявление шаблона выполняется блоком `{{define "template_name"}}...{{end}}`
- Имя шаблона указывается в `scheme.yaml`, в свойстве `template` и используется при генерации страницы конфигурации:
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
- Шаблоны могут вызывать другие шаблоны:
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
- Имя шаблона должно быть уникальным по сравнению с остальными,
  иначе шаблоны могут переписать друг-друга.
- В шаблоне можно использовать переменные для подстановки значений из `scheme.yaml`/LDAP:
  ```yaml
  # Все переменные хранятся в "точке":
  {{define "default"}}
  {{.var_name}}
  {{end}}
  ```
- Переменные объявляются  в `scheme.yaml`, в свойстве `fields`:
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
### Шаблоны + конфигурация для Thunderbird
Используя [PrefApi](#prefapi), подготавливается [шаблон](#общее) (например `default.tmpl`)

Основой для него может быть файл `pref.js` в каталоге **уже настроенного профиля** thunderbird:
- Windows - `%APPDATA%\Thunderbird\Profiles\*\prefs.js`
- Linux: - `~/.thunderbird/*/prefs.js`

> **Внимание!** 
> Изменение самого файла `pref.js` бесполезно, поскольку он обновляется почтовым клиентом в процессе работы.

# Thunderbird
Ниже будут **краткие** выжимки из различных источников, которые позволят подготовить
почтовый клиент к автоматической настройке.

Источники:
- [The SIPB Thunderbird Locker - Maintainers
](https://web.mit.edu/~thunderbird/www/maintainers/autoconfig.html)
- [MCD, Mission Control Desktop, AKA AutoConfig](https://udn.realityripple.com/docs/Archive/Misc_top_level/MCD,_Mission_Control_Desktop_AKA_AutoConfig)
- [EASY THUNDERBIRD ACCOUNT MANAGEMENT USING MCD](https://blog.deanandadie.net/2010/06/easy-thunderbird-account-management-using-mcd/)
- [Идеальный корпоративный почтовый клиент
](https://habr.com/ru/articles/101905/)
- [Customizing Firefox Using AutoConfig](https://support.mozilla.org/en-US/kb/customizing-firefox-using-autoconfig)

## PrefApi
\- это функции конфигурирования, управляющие параметрами почтового клиента при его запуске.

Определение функций PrefApi выполнено в файле `$THUNDERBIRD_FOLDER/omni.ja/defaults/autoconfig/prefcalls.js`:
- `getPrefBranch()` - получает корень дерева предпочтений. **Напрямую не используется**.
- `pref(prefName, value)` - меняет текущее значение на указанное. **Используется наиболее часто**.
- `defaultPref(prefName, value)` - устанавливает для параметра значение по умолчанию.
    Т.е. если пользователь оставляет поле пустым или выполняет сброс настроек, будет выставлено указанное значение.
- `lockPref(prefName, value)` - блокирует параметр на указанном значении.
    Пользователи не могут изменить заблокированные параметры.
- `unlockPref(prefName)` - разблокирует, ранее была заблокированный, параметр.
- `getPref(prefName)` - получает значение указанного параметра.
- `displayError(funcname, message)`- выводит диалоговое окно с сообщением об ошибке при запуске Thunderbird. **Единственное средство отладки переменных**.
- `getenv(name)` - получает значение указанной переменной среды из среды пользователя.
- Функции работы с LDAP. Не поддерживают авторизацию, поэтому бесполезны:
  - `setLDAPVersion(version)` - устанавливает версию LDAP, используемую сервером.
  - `getLDAPAttributes(host, base, filter, attribs)` - получает атрибуты LDAP с данного сервера.
  - `getLDAPValue(str, key)` - получает значение LDAP для заданной строки, отфильтрованное по ключу.

## Файлы конфигурации
### thunderbird.cfg
Файл настроек представляет из себя JavaScript, поэтому ему доступны переменные, функции и т.д.
Для использования внешнего источника настроек, пишется скрипт `thunderbird.cfg`:
```js
// Загрузка настроек с сервера
try {
    // Получение имени пользователя из ОС
    if (getenv("USER") != "") {
        // *NIX settings
        var env_user = getenv("USER");
    } else {
        // Windows settings
        var env_user = getenv("USERNAME");
    }
// Указание сервера конфигурации
pref("autoadmin.global_config_url", "http://server_host/"+env_user);
// Не добавлять переменную почты в запрос
pref("autoadmin.append_emailaddr", false);
 
} catch (e) {
    displayError("pref", e);
}
```
 и кладётся в каталог программы:
- Windows - `C:\Program Files (x86)\Mozilla Thunderbird\thunderbird.cfg`
- Linux - `/usr/lib/thunderbird/thunderbird.cfg`

### autoconfig.js
Для того, чтобы `thunderbird.cfg` был задействован, необходимо создать файл `autoconfig.js`:
```js
// Указываю файл конфигурации
pref('general.config.filename', 'thunderbird.cfg');
// Отключаю битовое смещение,чтобы читать обычный файл
pref('general.config.obscure_value', 0);
```
и кладу его по пути:
- Windows - `C:\Program Files\Mozilla Thunderbird\defaults\pref\autoconf.js`
- Linux - `/usr/share/thunderbird/defaults/pref/autoconfig.js`


# Безопасность
Код проверен статическими анализаторами, которые ориентированы на поиск уязвимостей.

## GoSec
Установка:
```sh
go install github.com/securego/gosec/v2/cmd/gosec@latest
```
Запуск:
```sh
gosec ./...
```
## Go Vulnerability
Установка:
```sh
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Запуск:
```sh
govulncheck ./...
```
