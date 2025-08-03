# API Documentation

This documentation provides details about the API endpoints available for the API Visualization and Analytics tool.

## Endpoints

- [POST /api/visualize](#post-apivisualize)
- [GET /api/stats](#get-apistats)
- [GET /api/history](#get-apihistory)

---

## POST /api/visualize

This endpoint is used to record a single API request for visualization and analytics. It accepts a JSON payload with the details of the request.

### Request Body

| Field          | Type    | Description                               |
|----------------|---------|-------------------------------------------|
| `endpoint`     | string  | The API endpoint path (e.g., `/users`).   |
| `method`       | string  | The HTTP method (e.g., `GET`, `POST`).    |
| `status_code`  | integer | The HTTP status code of the response.     |
| `response_time`| float64 | The response time in milliseconds.        |
| `error`        | string  | Any error message, if an error occurred.  |
| `timestamp`    | string  | The timestamp of the request in RFC3339 format. |

**Example:**

```json
{
  "endpoint": "/users",
  "method": "GET",
  "status_code": 200,
  "response_time": 123.45,
  "error": "",
  "timestamp": "2025-08-03T10:00:00Z"
}
```

### Responses

- **200 OK**: The request was successfully recorded.
- **400 Bad Request**: The request body is invalid.

---

## GET /api/stats

This endpoint retrieves real-time aggregated statistics for all recorded routes. The statistics are sorted by the number of requests in descending order.

### Responses

- **200 OK**: Returns a JSON array of `RouteStats` objects.

  ```json
  [
    {
      "path": "/users",
      "method": "GET",
      "count": 100,
      "last_accessed": "2025-08-03T10:30:00Z",
      "avg_response": 123456789,
      "status_codes": {
        "200": 100
      },
      "errors": []
    }
  ]
  ```

---

## GET /api/history

This endpoint retrieves historical request data for detailed analytics. It supports filtering by time range, endpoint, and method.

### Query Parameters

| Parameter   | Type   | Description                                                                                                 | Default |
|-------------|--------|-------------------------------------------------------------------------------------------------------------|---------|
| `timeRange` | string | The time range to fetch data for. Supported values: `1h`, `6h`, `24h`, `168h` (7 days).                        | `1h`    |
| `endpoint`  | string | The specific endpoint to filter by (e.g., `/users`). If not provided or set to `all`, all endpoints are included. | `all`   |
| `method`    | string | The HTTP method to filter by (e.g., `GET`). If not provided or set to `all`, all methods are included.        | `all`   |
| `month`     | string | The month to fetch data for, in `YYYY-MM` format. This overrides `timeRange`.                                 | -       |

### Responses

- **200 OK**: Returns a JSON array of `VisualizationData` objects.

  ```json
  [
    {
      "endpoint": "/users",
      "method": "GET",
      "status_code": 200,
      "response_time": 123.45,
      "error": "",
      "timestamp": "2025-08-03T10:00:00Z"
    }
  ]
  ```

- **500 Internal Server Error**: If there is a database query error.
