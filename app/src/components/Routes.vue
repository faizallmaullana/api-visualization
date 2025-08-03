<template>
  <div class="routes-section">
    <!-- Routes Filters -->
    <div class="analytics-filters">
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Method Filter</label>
        <select v-model="filters.method" @change="loadRoutesData" class="analytics-select">
          <option value="all">All Methods</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
          <option value="PATCH">PATCH</option>
        </select>
      </div>
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Time Range</label>
        <select v-model="filters.timeRange" @change="loadRoutesData" class="analytics-select">
          <option value="1h">Last 1 Hour</option>
          <option value="6h">Last 6 Hours</option>
          <option value="24h">Last 24 Hours</option>
          <option value="168h">Last 7 Days</option>
        </select>
      </div>
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Sort By</label>
        <select v-model="filters.sortBy" @change="updateRoutesList" class="analytics-select">
          <option value="requests">Total Requests</option>
          <option value="response_time">Response Time</option>
          <option value="errors">Error Count</option>
          <option value="endpoint">Endpoint Name</option>
        </select>
      </div>
      <button @click="loadRoutesData" class="analytics-button">
        <i class="fas fa-sync-alt"></i> Refresh
      </button>
    </div>

    <!-- Routes Overview Stats -->
    <div class="dashboard-grid">
      <div class="panel col-3">
        <div class="panel-header">
          <div class="panel-title">Total Endpoints</div>
        </div>
        <div class="panel-content">
          <div class="metric-card">
            <div class="metric-value">{{ routeMetrics.totalEndpoints }}</div>
          </div>
        </div>
      </div>
      <div class="panel col-3">
        <div class="panel-header">
          <div class="panel-title">Most Active</div>
        </div>
        <div class="panel-content">
          <div class="metric-card">
            <div class="metric-value">{{ routeMetrics.mostActive }}</div>
          </div>
        </div>
      </div>
      <div class="panel col-3">
        <div class="panel-header">
          <div class="panel-title">Slowest Route</div>
        </div>
        <div class="panel-content">
          <div class="metric-card">
            <div class="metric-value">{{ routeMetrics.slowestRoute }}</div>
          </div>
        </div>
      </div>
      <div class="panel col-3">
        <div class="panel-header">
          <div class="panel-title">Error Rate</div>
        </div>
        <div class="panel-content">
          <div class="metric-card">
            <div class="metric-value">{{ routeMetrics.errorRate }}%</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Routes Charts Container -->
    <div class="routes-charts-container">
      <div v-for="route in sortedRoutes" :key="route.endpoint + route.method" class="routes-endpoint-chart">
        <div class="routes-endpoint-header">
          <div class="routes-endpoint-title">{{ route.endpoint }}</div>
          <div class="routes-endpoint-method" :class="'method-' + route.method">{{ route.method }}</div>
        </div>
        
        <div class="routes-endpoint-stats">
          <div class="routes-stat-item">
            <div class="routes-stat-value">{{ route.count }}</div>
            <div class="routes-stat-label">Requests</div>
          </div>
          <div class="routes-stat-item">
            <div class="routes-stat-value">{{ route.avgResponseTime }}ms</div>
            <div class="routes-stat-label">Avg Time</div>
          </div>
          <div class="routes-stat-item">
            <div class="routes-stat-value">{{ route.errorCount }}</div>
            <div class="routes-stat-label">Errors</div>
          </div>
          <div class="routes-stat-item">
            <div class="routes-stat-value">{{ route.successRate }}%</div>
            <div class="routes-stat-label">Success</div>
          </div>
        </div>
        
        <div class="routes-chart-container">
          <canvas :ref="'route-chart-' + route.id" :id="'route-chart-' + route.id"></canvas>
        </div>
        
        <div v-if="route.errors.length" class="routes-error-list">
          <div class="routes-error-title">Recent Errors</div>
          <div v-for="(error, i) in route.errors.slice(0, 5)" :key="i" class="routes-error-item">
            {{ error }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue';
import Chart from 'chart.js/auto';
import { useApi } from '../service/api';

const historyData = ref([]);
const filters = ref({
  method: 'all',
  timeRange: '24h',
  sortBy: 'requests'
});
const routeCharts = ref({});
const api = useApi();

// Process history data into route statistics
const processedRoutes = computed(() => {
  const routes = {};
  
  historyData.value.forEach(d => {
    const key = `${d.method} ${d.endpoint}`;
    if (!routes[key]) {
      routes[key] = {
        id: key.replace(/\s+/g, '-').replace(/[^a-zA-Z0-9-]/g, ''),
        endpoint: d.endpoint,
        method: d.method,
        requests: [],
        errors: []
      };
    }
    
    routes[key].requests.push(d);
    if (d.status_code >= 400 && d.error) {
      routes[key].errors.push(d.error);
    }
  });
  
  return Object.values(routes).map(route => {
    const count = route.requests.length;
    const avgResponseTime = count ? (route.requests.reduce((sum, r) => sum + r.response_time, 0) / count).toFixed(1) : 0;
    const errorCount = route.errors.length;
    const successRate = count ? (((count - errorCount) / count) * 100).toFixed(1) : 0;
    
    return {
      ...route,
      count,
      avgResponseTime: parseFloat(avgResponseTime),
      errorCount,
      successRate: parseFloat(successRate)
    };
  });
});

const sortedRoutes = computed(() => {
  const routes = [...processedRoutes.value];
  const sortBy = filters.value.sortBy;
  
  return routes.sort((a, b) => {
    switch (sortBy) {
      case 'requests': return b.count - a.count;
      case 'response_time': return b.avgResponseTime - a.avgResponseTime;
      case 'errors': return b.errorCount - a.errorCount;
      case 'endpoint': return a.endpoint.localeCompare(b.endpoint);
      default: return 0;
    }
  });
});

const routeMetrics = computed(() => {
  const routes = processedRoutes.value;
  if (!routes.length) return { totalEndpoints: 0, mostActive: 'N/A', slowestRoute: 'N/A', errorRate: 0 };
  
  const totalEndpoints = routes.length;
  const mostActive = routes.reduce((max, r) => r.count > max.count ? r : max, routes[0]);
  const slowestRoute = routes.reduce((max, r) => r.avgResponseTime > max.avgResponseTime ? r : max, routes[0]);
  const totalRequests = routes.reduce((sum, r) => sum + r.count, 0);
  const totalErrors = routes.reduce((sum, r) => sum + r.errorCount, 0);
  const errorRate = totalRequests ? ((totalErrors / totalRequests) * 100).toFixed(1) : 0;
  
  return {
    totalEndpoints,
    mostActive: `${mostActive.method} ${mostActive.endpoint}`,
    slowestRoute: `${slowestRoute.avgResponseTime}ms`,
    errorRate
  };
});

const loadRoutesData = async () => {
  try {
    const params = {
      timeRange: filters.value.timeRange
    };
    if (filters.value.method !== 'all') params.method = filters.value.method;
    
    const response = await api.getHistory(params);
    historyData.value = response.data || [];
    
    await nextTick();
    updateRouteCharts();
  } catch (error) {
    console.error('Failed to load routes data:', error);
    historyData.value = [];
  }
};

const updateRoutesList = () => {
  // Trigger reactive update
};

const updateRouteCharts = () => {
  sortedRoutes.value.forEach(route => {
    const canvasId = `route-chart-${route.id}`;
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    
    // Destroy existing chart
    if (routeCharts.value[route.id]) {
      routeCharts.value[route.id].destroy();
    }
    
    // Generate complete time range based on filter
    const now = new Date();
    const timeRange = filters.value.timeRange;
    let minutes = 60; // default 1 hour
    switch(timeRange) {
      case '1h': minutes = 60; break;
      case '6h': minutes = 360; break;
      case '24h': minutes = 1440; break;
      case '168h': minutes = 10080; break; // 7 days
    }
    
    // Create complete time buckets with zero values
    const timeLabels = [];
    const timeDataMap = {};
    
    for (let i = minutes - 1; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 1000); // minutes ago
      const timeKey = time.toISOString().slice(0, 16) + ':00.000Z';
      timeLabels.push(time);
      timeDataMap[timeKey] = 0;
    }
    
    // Fill with actual data
    route.requests.forEach(r => {
      const timeKey = new Date(r.timestamp || Date.now()).toISOString().slice(0, 16) + ':00.000Z';
      if (timeDataMap.hasOwnProperty(timeKey)) {
        timeDataMap[timeKey] = r.response_time;
      }
    });
    
    // Prepare chart data with zeros for missing periods
    const chartData = timeLabels.map(time => ({
      x: time,
      y: timeDataMap[time.toISOString().slice(0, 16) + ':00.000Z'] || 0
    }));
    
    // Create chart
    const ctx = canvas.getContext('2d');
    routeCharts.value[route.id] = new Chart(ctx, {
      type: 'line',
      data: {
        datasets: [{
          label: 'Response Time (ms)',
          data: chartData,
          borderColor: '#52c5f7',
          backgroundColor: 'rgba(82, 197, 247, 0.1)',
          fill: true,
          tension: 0.1,
          spanGaps: false,
          pointRadius: 1
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          x: {
            type: 'time',
            time: { unit: 'minute' },
            grid: { color: '#262626' }
          },
          y: {
            beginAtZero: true,
            grid: { color: '#262626' }
          }
        },
        plugins: {
          legend: { display: false }
        }
      }
    });
  });
};

onMounted(() => {
  loadRoutesData();
});
</script>

<style scoped>
.analytics-filters {
  background: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 20px;
  display: flex;
  gap: 15px;
  align-items: center;
  flex-wrap: wrap;
}

.analytics-filter-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.analytics-filter-label {
  font-size: 12px;
  color: var(--grafana-text-muted);
}

.analytics-select {
  background: var(--grafana-bg);
  border: 1px solid var(--grafana-border);
  color: var(--grafana-text);
  padding: 6px 10px;
  border-radius: 3px;
  font-size: 13px;
}

.analytics-button {
  background: var(--grafana-blue);
  color: #000;
  border: none;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
}

.metric-card {
  text-align: center;
}

.metric-value {
  font-size: 24px;
  font-weight: 500;
  color: var(--grafana-text);
}

.routes-charts-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
  gap: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.routes-endpoint-chart {
  background: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 4px;
  padding: 15px;
  max-width: 600px;
}

.routes-endpoint-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  border-bottom: 1px solid var(--grafana-border);
  padding-bottom: 10px;
}

.routes-endpoint-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--grafana-text);
}

.routes-endpoint-method {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.method-GET { background-color: var(--grafana-blue); color: #000; }
.method-POST { background-color: var(--grafana-green); color: #000; }
.method-PUT { background-color: var(--grafana-orange); color: #000; }
.method-DELETE { background-color: var(--grafana-red); color: #fff; }
.method-PATCH { background-color: var(--grafana-purple); color: #fff; }

.routes-endpoint-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 10px;
  margin-bottom: 15px;
  padding: 10px;
  background: var(--grafana-bg-light);
  border-radius: 4px;
}

.routes-stat-item {
  text-align: center;
}

.routes-stat-value {
  font-size: 18px;
  font-weight: 500;
  color: var(--grafana-text);
  margin-bottom: 2px;
}

.routes-stat-label {
  font-size: 10px;
  color: var(--grafana-text-muted);
  text-transform: uppercase;
}

.routes-chart-container {
  height: 180px;
  width: 100%;
  max-width: 500px;
  margin: 0 auto 15px auto;
}

.routes-error-list {
  padding: 10px;
  background: var(--grafana-bg-light);
  border-radius: 4px;
  border-left: 3px solid var(--grafana-red);
}

.routes-error-title {
  font-size: 12px;
  color: var(--grafana-red);
  font-weight: 500;
  margin-bottom: 8px;
}

.routes-error-item {
  font-size: 11px;
  color: var(--grafana-text-muted);
  margin-bottom: 4px;
  padding-left: 10px;
}
</style>
