


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

###  save_c_s_v

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| POST | /api/v1/save_csv | [save csv](#save-csv) | Принимает csv файл и сохраняет информацию в БД |
  


## Paths

### <span id="save-csv"></span> Принимает csv файл и сохраняет информацию в БД (*save csv*)

```
POST /api/v1/save_csv
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


