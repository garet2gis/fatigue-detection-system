


# User data service API
  

## Informations

### Version

1.0

### Contact

  

## Content negotiation

### URI Schemes
  * http

### Consumes
  * application/json

### Produces
  * application/json

## All endpoints

###  auth

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/v1/auth/login | [login](#login) | Принимает данные пользователя для входа в систему |
| POST | /api/v1/auth/register | [register](#register) | Принимает данные пользователя |
  


###  save_c_s_v

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/v1/face_model/save_features | [save csv](#save-csv) | Принимает csv файл с фичами из видео |
  


## Paths

### <span id="login"></span> Принимает данные пользователя для входа в систему (*login*)

```
POST /api/v1/auth/login
```

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_credentials | `body` | [LoginBody](#login-body) | `LoginBody` | | ✓ | | Данные для логина |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#login-200) | OK | OK |  | [schema](#login-200-schema) |
| [400](#login-400) | Bad Request | Bad Request |  | [schema](#login-400-schema) |

#### Responses


##### <span id="login-200"></span> 200 - OK
Status: OK

###### <span id="login-200-schema"></span> Schema
   
  

[LoginOKBody](#login-o-k-body)

##### <span id="login-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="login-400-schema"></span> Schema
   
  

[LoginBadRequestBody](#login-bad-request-body)

###### Inlined models

**<span id="login-bad-request-body"></span> LoginBadRequestBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | integer| `int64` | ✓ | | Код ошибки | `26002` |
| message | string| `string` | ✓ | | Сообщение ошибки | `entity not found` |
| name | string| `string` | ✓ | | Наименование ошибки | `NotFound` |
| status | integer| `int64` | ✓ | | Статус код ответа | `404` |



**<span id="login-body"></span> LoginBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| login | string| `string` | ✓ | |  |  |
| password | string| `string` | ✓ | |  |  |



**<span id="login-o-k-body"></span> LoginOKBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| get_face_model_url | string| `string` |  | |  |  |
| upload_face_model_features_url | string| `string` |  | |  |  |



### <span id="register"></span> Принимает данные пользователя (*register*)

```
POST /api/v1/auth/register
```

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| user_data | `body` | [RegisterBody](#register-body) | `RegisterBody` | | ✓ | | Данные для регистрации |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [204](#register-204) | No Content | No Content |  | [schema](#register-204-schema) |
| [400](#register-400) | Bad Request | Bad Request |  | [schema](#register-400-schema) |

#### Responses


##### <span id="register-204"></span> 204 - No Content
Status: No Content

###### <span id="register-204-schema"></span> Schema

##### <span id="register-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="register-400-schema"></span> Schema
   
  

[RegisterBadRequestBody](#register-bad-request-body)

###### Inlined models

**<span id="register-bad-request-body"></span> RegisterBadRequestBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | integer| `int64` | ✓ | | Код ошибки | `26002` |
| message | string| `string` | ✓ | | Сообщение ошибки | `entity not found` |
| name | string| `string` | ✓ | | Наименование ошибки | `NotFound` |
| status | integer| `int64` | ✓ | | Статус код ответа | `404` |



**<span id="register-body"></span> RegisterBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| login | string| `string` | ✓ | |  |  |
| name | string| `string` |  | |  |  |
| password | string| `string` | ✓ | |  |  |
| surname | string| `string` |  | |  |  |



### <span id="save-csv"></span> Принимает csv файл с фичами из видео (*save csv*)

```
POST /api/v1/face_model/save_features
```

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| file | `formData` | file | `io.ReadCloser` |  | ✓ |  | Загружаемый csv |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [204](#save-csv-204) | No Content | No Content |  | [schema](#save-csv-204-schema) |
| [400](#save-csv-400) | Bad Request | Bad Request |  | [schema](#save-csv-400-schema) |

#### Responses


##### <span id="save-csv-204"></span> 204 - No Content
Status: No Content

###### <span id="save-csv-204-schema"></span> Schema

##### <span id="save-csv-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="save-csv-400-schema"></span> Schema
   
  

[SaveCsvBadRequestBody](#save-csv-bad-request-body)

###### Inlined models

**<span id="save-csv-bad-request-body"></span> SaveCsvBadRequestBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | integer| `int64` | ✓ | | Код ошибки | `26002` |
| message | string| `string` | ✓ | | Сообщение ошибки | `entity not found` |
| name | string| `string` | ✓ | | Наименование ошибки | `NotFound` |
| status | integer| `int64` | ✓ | | Статус код ответа | `404` |



## Models

### <span id="app-errors-app-error"></span> app_errors.AppError


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | integer| `int64` | ✓ | | Код ошибки | `26002` |
| message | string| `string` | ✓ | | Сообщение ошибки | `entity not found` |
| name | string| `string` | ✓ | | Наименование ошибки | `NotFound` |
| status | integer| `int64` | ✓ | | Статус код ответа | `404` |



### <span id="fixtures-login-request"></span> fixtures.LoginRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| login | string| `string` | ✓ | |  |  |
| password | string| `string` | ✓ | |  |  |



### <span id="fixtures-login-response"></span> fixtures.LoginResponse


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| get_face_model_url | string| `string` |  | |  |  |
| upload_face_model_features_url | string| `string` |  | |  |  |



### <span id="fixtures-register-request"></span> fixtures.RegisterRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| login | string| `string` | ✓ | |  |  |
| name | string| `string` |  | |  |  |
| password | string| `string` | ✓ | |  |  |
| surname | string| `string` |  | |  |  |


