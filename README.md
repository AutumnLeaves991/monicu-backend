# mon.icu backend

This repository contains code of the backend software used on the server side of mon.icu project.

Currently, the project is not in a stable phase and undergoes huge refactorings and enhancements that are most likely
breaking.

## API

Pre-release public API is already available at base URL `https://api.mon.icu`.

### Endpoints

#### GET /posts/:page

##### URL parameters

|Name|Type                   |Required|Example|
|----|-----------------------|--------|-------|
|page|unsigned 32-bit integer|✘       |42     |

##### Responses

##### 200 OK

Example response body (JSON, prettified):

```json
[
  {
    "id": 1,
    "channel": 1,
    "user": 1,
    "images": [
      {
        "url": "https://example.com/image.jpg",
        "width": 800,
        "height": 800,
        "size": 640000
      }
    ],
    "reactions": 1
  }
]
```

##### 400 Bad Request

Example response body (JSON, prettified):

```json
{
  "error": "An error occurred."
}
```

##### 503 Internal Server Error

Content is exactly the same as in [400 Bad Request](#400-bad-request).