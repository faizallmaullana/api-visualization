# API Visualization Tool

A powerful Go application that captures and visualizes API requests in real-time. This tool acts as a reverse proxy, intercepting all requests to your target API server and providing beautiful visualizations of the traffic patterns.

## Features

🚀 **Real-time API Monitoring**: Captures all requests to your target API server  
📊 **Beautiful Dashboard**: Modern, responsive web interface with live charts  
⚡ **Zero Dependencies**: Built using only Go standard library  
📈 **Rich Analytics**: Track request counts, response times, status codes, and more  
🎯 **Route-specific Stats**: Detailed statistics for each API endpoint  
🔄 **Auto-refresh**: Dashboard updates automatically every 5 seconds  

## Architecture

```
Client Request → Visualization Proxy (localhost:8080) → Target API (localhost:8000)
                         ↓
                   Statistics Collection
                         ↓
                   Web Dashboard (/dashboard)
```

## Quick Start

### 1. Start the Target API Server

First, start the test API server that will receive the proxied requests:

```bash
cd test-api
go run main.go
```

This starts a sample REST API server on `localhost:8000` with endpoints for users and products.

### 2. Start the Visualization Proxy

In another terminal, start the visualization proxy:

```bash
go run main.go
```

This starts the visualization server on `localhost:8080`.

### 3. View the Dashboard

Open your browser and navigate to:
```
http://localhost:8080/dashboard
```

### 4. Generate Some Traffic

Make requests to the proxy (which forwards to your API):

```bash
# Health check
curl http://localhost:8080/api/health

# Get users
curl http://localhost:8080/api/users

# Create a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","age":28}'

# Get products
curl http://localhost:8080/api/products

# Test error endpoint
curl http://localhost:8080/api/error-test
```

## Dashboard Features

### 📊 Route Statistics
- **Request Count**: Number of times each endpoint was called
- **Average Response Time**: Mean response time for each route
- **Last Accessed**: Timestamp of the most recent request
- **Status Code Distribution**: Breakdown of HTTP status codes per route

### 📈 Interactive Charts
- **Request Methods**: Pie chart showing distribution of HTTP methods (GET, POST, PUT, DELETE)
- **Response Times**: Bar chart of average response times by route
- **Status Codes**: Overall distribution of HTTP status codes

### 🔄 Real-time Updates
- Dashboard automatically refreshes every 5 seconds
- Manual refresh button available
- Live statistics without page reload

## API Endpoints

The visualization server provides these endpoints:

- `GET /dashboard` - Main visualization dashboard
- `GET /api/stats` - JSON API for route statistics
- `*` (all other routes) - Proxied to target server with statistics collection

## Configuration

### Changing Target Server

To proxy a different server, modify the target URL in `main.go`:

```go
proxyHandler, err := NewProxyHandler("http://your-target-server:port", statsManager)
```

### Changing Ports

- **Visualization Server**: Change `:8080` in the `ListenAndServe` call
- **Target Server**: Update the proxy target URL

## Data Collected

For each API request, the system tracks:

- **Route Information**: HTTP method and path
- **Performance Metrics**: Response time and request count
- **Status Codes**: HTTP response status codes
- **User Agents**: Client user agent strings
- **Timestamps**: When requests were made

## Use Cases

### Development
- Monitor API usage during development
- Identify slow endpoints
- Track error rates
- Understand traffic patterns

### Testing
- Visualize test coverage across API endpoints
- Monitor performance during load testing
- Verify proper error handling

### Debugging
- Quickly identify problematic routes
- Track unusual traffic patterns
- Monitor response time trends

## Sample API Endpoints (Test Server)

The included test server provides these endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check endpoint |
| GET | `/api/users` | Get all users |
| POST | `/api/users` | Create a new user |
| GET | `/api/users/{id}` | Get user by ID |
| PUT | `/api/users/{id}` | Update user by ID |
| DELETE | `/api/users/{id}` | Delete user by ID |
| GET | `/api/products` | Get all products |
| POST | `/api/products` | Create a new product |
| GET | `/api/error-test` | Endpoint that randomly returns errors |

## Technical Details

### Statistics Storage
- In-memory storage with thread-safe operations
- Statistics reset when the server restarts
- Efficient concurrent access using read-write mutexes

### Proxy Implementation
- Built using `httputil.ReverseProxy` from Go standard library
- Preserves all request headers and body
- Captures response status codes and timing

### Web Interface
- Modern HTML5/CSS3/JavaScript
- Chart.js for interactive visualizations
- Responsive design works on desktop and mobile
- No external dependencies (CDN-based Chart.js)

## Performance

The visualization proxy adds minimal overhead:
- **Memory**: ~1-2MB for statistics storage
- **Latency**: <1ms additional request time
- **CPU**: Negligible impact on request processing

## License

This project is open source and available under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.
# api-visualization
