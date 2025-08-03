
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

  ```json
  "DB query error"
  ```
