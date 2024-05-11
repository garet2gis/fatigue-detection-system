


# Model storage service
  

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

###  features

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/v1/increase_features | [increase features](#increase-features) | Принимает количество новых фич по моделям |
  


###  models

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/v1/get_models | [get models](#get-models) | Возвращает ссылки на модели по id пользователя |
| POST | /api/v1/save_model | [save model](#save-model) | Принимает ml модель |
  


## Paths

### <span id="get-models"></span> Возвращает ссылки на модели по id пользователя (*get models*)

```
POST /api/v1/get_models
```

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| features_data | `body` | [GetModelsBody](#get-models-body) | `GetModelsBody` | | ✓ | | ID пользователя |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#get-models-200) | OK | OK |  | [schema](#get-models-200-schema) |
| [400](#get-models-400) | Bad Request | Bad Request |  | [schema](#get-models-400-schema) |

#### Responses


##### <span id="get-models-200"></span> 200 - OK
Status: OK

###### <span id="get-models-200-schema"></span> Schema
   
  

map of string

##### <span id="get-models-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="get-models-400-schema"></span> Schema
   
  

[GetModelsBadRequestBody](#get-models-bad-request-body)

###### Inlined models

**<span id="get-models-bad-request-body"></span> GetModelsBadRequestBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | integer| `int64` | ✓ | | Код ошибки | `26002` |
| message | string| `string` | ✓ | | Сообщение ошибки | `entity not found` |
| name | string| `string` | ✓ | | Наименование ошибки | `NotFound` |
| status | integer| `int64` | ✓ | | Статус код ответа | `404` |



**<span id="get-models-body"></span> GetModelsBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| string | string| `string` | ✓ | |  |  |



### <span id="increase-features"></span> Принимает количество новых фич по моделям (*increase features*)

```
POST /api/v1/increase_features
```

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| features_data | `body` | [IncreaseFeaturesBody](#increase-features-body) | `IncreaseFeaturesBody` | | ✓ | | Данные о количестве фич |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [204](#increase-features-204) | No Content | No Content |  | [schema](#increase-features-204-schema) |
| [400](#increase-features-400) | Bad Request | Bad Request |  | [schema](#increase-features-400-schema) |

#### Responses


##### <span id="increase-features-204"></span> 204 - No Content
Status: No Content

###### <span id="increase-features-204-schema"></span> Schema

##### <span id="increase-features-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="increase-features-400-schema"></span> Schema
   
  

[IncreaseFeaturesBadRequestBody](#increase-features-bad-request-body)

###### Inlined models

**<span id="increase-features-bad-request-body"></span> IncreaseFeaturesBadRequestBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| code | integer| `int64` | ✓ | | Код ошибки | `26002` |
| message | string| `string` | ✓ | | Сообщение ошибки | `entity not found` |
| name | string| `string` | ✓ | | Наименование ошибки | `NotFound` |
| status | integer| `int64` | ✓ | | Статус код ответа | `404` |



**<span id="increase-features-body"></span> IncreaseFeaturesBody**


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| features_count | integer| `int64` | ✓ | |  |  |
| model_type | string| `string` | ✓ | |  |  |
| user_id | string| `string` | ✓ | |  |  |



### <span id="save-model"></span> Принимает ml модель (*save model*)

```
POST /api/v1/save_model
```

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| file | `formData` | file | `io.ReadCloser` |  | ✓ |  | Загружаемая ml-модель |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [204](#save-model-204) | No Content | No Content |  | [schema](#save-model-204-schema) |
| [400](#save-model-400) | Bad Request | Bad Request |  | [schema](#save-model-400-schema) |

#### Responses


##### <span id="save-model-204"></span> 204 - No Content
Status: No Content

###### <span id="save-model-204-schema"></span> Schema

##### <span id="save-model-400"></span> 400 - Bad Request
Status: Bad Request

###### <span id="save-model-400-schema"></span> Schema
   
  

[SaveModelBadRequestBody](#save-model-bad-request-body)

###### Inlined models

**<span id="save-model-bad-request-body"></span> SaveModelBadRequestBody**


  



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



### <span id="fixtures-get-models-request"></span> fixtures.GetModelsRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| string | string| `string` | ✓ | |  |  |



### <span id="fixtures-increase-features-request"></span> fixtures.IncreaseFeaturesRequest


  



**Properties**

| Name | Type | Go type | Required | Default | Description | Example |
|------|------|---------|:--------:| ------- |-------------|---------|
| features_count | integer| `int64` | ✓ | |  |  |
| model_type | string| `string` | ✓ | |  |  |
| user_id | string| `string` | ✓ | |  |  |


