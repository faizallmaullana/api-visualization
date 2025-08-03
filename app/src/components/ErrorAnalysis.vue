<template>
  <div class="error-analysis-section">
    <!-- Page Header -->
    <PageHeader
      page-title="Error Analysis Dashboard"
      page-description="Comprehensive error tracking, pattern analysis, and impact assessment"
      page-icon="fas fa-exclamation-triangle"
      :stats="headerStats"
      :refreshing="isRefreshing"
      :show-export="true"
      @refresh="handleRefresh"
      @export="handleExport"
    >
      <template #filters>
        <div class="filter-group">
          <label class="filter-label">Time Range</label>
          <select class="filter-select" v-model="filters.timeRange" @change="loadErrorData">
            <option value="1h">Last 1 Hour</option>
            <option value="6h">Last 6 Hours</option>
            <option value="24h">Last 24 Hours</option>
            <option value="168h">Last 7 Days</option>
            <option value="720h">Last 30 Days</option>
          </select>
        </div>
        <div class="filter-group">
          <label class="filter-label">Error Type</label>
          <select class="filter-select" v-model="filters.errorType" @change="loadErrorData">
            <option value="all">All Errors</option>
            <option value="4xx">4xx Client Errors</option>
            <option value="5xx">5xx Server Errors</option>
          </select>
        </div>
        <div class="filter-group">
          <label class="filter-label">Status Code</label>
          <select class="filter-select" v-model="filters.statusCode" @change="loadErrorData">
            <option value="all">All Status Codes</option>
            <option v-for="code in uniqueStatusCodes" :key="code" :value="code">{{ code }}</option>
          </select>
        </div>
        <div class="filter-group">
          <label class="filter-label">Endpoint</label>
          <select class="filter-select" v-model="filters.endpoint" @change="loadErrorData">
            <option value="all">All Endpoints</option>
            <option v-for="ep in uniqueEndpoints" :key="ep" :value="ep">{{ ep }}</option>
          </select>
        </div>
      </template>
    </PageHeader>

    <!-- Error Summary Stats -->
    <div class="error-summary">
      <div class="summary-card error">
        <div class="card-header">
          <i class="fas fa-exclamation-triangle"></i>
          <span>Total Errors</span>
        </div>
        <div class="card-value">{{ errorSummary.totalErrors }}</div>
        <div class="card-subtitle">{{ errorSummary.errorRate }}% of total requests</div>
      </div>
      <div class="summary-card client">
        <div class="card-header">
          <i class="fas fa-user-times"></i>
          <span>4xx Client Errors</span>
        </div>
        <div class="card-value">{{ errorSummary.clientErrors }}</div>
        <div class="card-subtitle">{{ errorSummary.clientErrorRate }}% of errors</div>
      </div>
      <div class="summary-card server">
        <div class="card-header">
          <i class="fas fa-server"></i>
          <span>5xx Server Errors</span>
        </div>
        <div class="card-value">{{ errorSummary.serverErrors }}</div>
        <div class="card-subtitle">{{ errorSummary.serverErrorRate }}% of errors</div>
      </div>
      <div class="summary-card frequent">
        <div class="card-header">
          <i class="fas fa-chart-line"></i>
          <span>Most Frequent</span>
        </div>
        <div class="card-value">{{ errorSummary.mostFrequentStatus }}</div>
        <div class="card-subtitle">{{ errorSummary.mostFrequentCount }} occurrences</div>
      </div>
    </div>

    <!-- Error Analysis Grid -->
    <div class="error-grid">
      <!-- Error Timeline -->
      <div class="panel col-12">
        <div class="panel-header">
          <div class="panel-title">Error Frequency Over Time</div>
          <div class="panel-subtitle">Breakdown by error type (4xx vs 5xx)</div>
        </div>
        <div class="panel-content">
          <div class="chart-container large-chart">
            <canvas ref="errorTimelineCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Error Distribution by Status Code -->
      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">Error Distribution by Status Code</div>
          <div class="panel-subtitle">Most common error codes</div>
        </div>
        <div class="panel-content">
          <div class="chart-container medium-chart">
            <canvas ref="statusDistributionCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Error Distribution by Endpoint -->
      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">Errors by Endpoint</div>
          <div class="panel-subtitle">Top 10 endpoints with most errors</div>
        </div>
        <div class="panel-content">
          <div class="chart-container medium-chart">
            <canvas ref="endpointErrorCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Error Rate by Hour -->
      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">Error Rate by Hour of Day</div>
          <div class="panel-subtitle">When do errors occur most frequently?</div>
        </div>
        <div class="panel-content">
          <div class="chart-container medium-chart">
            <canvas ref="hourlyErrorCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Response Time vs Errors -->
      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">Response Time vs Error Rate</div>
          <div class="panel-subtitle">Correlation between performance and errors</div>
        </div>
        <div class="panel-content">
          <div class="chart-container medium-chart">
            <canvas ref="responseTimeErrorCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Recent Error Details Table -->
      <div class="panel col-12">
        <div class="panel-header">
          <div class="panel-title">Recent Error Details</div>
          <div class="panel-subtitle">Latest {{ recentErrors.length }} errors with full context</div>
        </div>
        <div class="panel-content">
          <div class="error-table-container">
            <table class="error-table">
              <thead>
                <tr>
                  <th>Timestamp</th>
                  <th>Status</th>
                  <th>Method</th>
                  <th>Endpoint</th>
                  <th>Response Time</th>
                  <th>Error Type</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="error in recentErrors" :key="error.id" :class="getErrorRowClass(error.status_code)">
                  <td class="timestamp">{{ formatTimestamp(error.timestamp) }}</td>
                  <td class="status">
                    <span :class="getStatusClass(error.status_code)">{{ error.status_code }}</span>
                  </td>
                  <td class="method">
                    <span :class="getMethodClass(error.method)">{{ error.method }}</span>
                  </td>
                  <td class="endpoint">{{ error.endpoint }}</td>
                  <td class="response-time">{{ error.response_time }}ms</td>
                  <td class="error-type">{{ getErrorTypeDescription(error.status_code) }}</td>
                  <td class="actions">
                    <button class="action-btn" @click="showErrorDetails(error)" title="View Details">
                      <i class="fas fa-info-circle"></i>
                    </button>
                    <button class="action-btn" @click="copyErrorInfo(error)" title="Copy Info">
                      <i class="fas fa-copy"></i>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-if="!recentErrors.length" class="no-errors">
              <i class="fas fa-check-circle"></i>
              <p>No errors found in the selected time range</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Error Patterns Analysis -->
      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">Error Patterns</div>
          <div class="panel-subtitle">Common error scenarios</div>
        </div>
        <div class="panel-content">
          <div class="pattern-list">
            <div v-for="pattern in errorPatterns" :key="pattern.id" class="pattern-item">
              <div class="pattern-header">
                <span class="pattern-title">{{ pattern.title }}</span>
                <span class="pattern-count">{{ pattern.count }} occurrences</span>
              </div>
              <div class="pattern-description">{{ pattern.description }}</div>
              <div class="pattern-recommendation">
                <i class="fas fa-lightbulb"></i>
                {{ pattern.recommendation }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Error Impact Analysis -->
      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">Error Impact Analysis</div>
          <div class="panel-subtitle">Business impact assessment</div>
        </div>
        <div class="panel-content">
          <div class="impact-metrics">
            <div class="impact-item high">
              <div class="impact-label">High Impact Errors</div>
              <div class="impact-value">{{ errorImpact.highImpact }}</div>
              <div class="impact-desc">5xx errors affecting core functionality</div>
            </div>
            <div class="impact-item medium">
              <div class="impact-label">Medium Impact Errors</div>
              <div class="impact-value">{{ errorImpact.mediumImpact }}</div>
              <div class="impact-desc">4xx errors with high frequency</div>
            </div>
            <div class="impact-item low">
              <div class="impact-label">Low Impact Errors</div>
              <div class="impact-value">{{ errorImpact.lowImpact }}</div>
              <div class="impact-desc">Infrequent client errors</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error Detail Modal -->
    <div v-if="selectedError" class="error-modal-overlay" @click="closeErrorDetails">
      <div class="error-modal" @click.stop>
        <div class="modal-header">
          <h3>Error Details</h3>
          <button class="close-btn" @click="closeErrorDetails">&times;</button>
        </div>
        <div class="modal-content">
          <div class="error-detail-grid">
            <div class="detail-item">
              <label>Timestamp:</label>
              <span>{{ formatTimestamp(selectedError.timestamp) }}</span>
            </div>
            <div class="detail-item">
              <label>Status Code:</label>
              <span :class="getStatusClass(selectedError.status_code)">{{ selectedError.status_code }}</span>
            </div>
            <div class="detail-item">
              <label>HTTP Method:</label>
              <span>{{ selectedError.method }}</span>
            </div>
            <div class="detail-item">
              <label>Endpoint:</label>
              <span>{{ selectedError.endpoint }}</span>
            </div>
            <div class="detail-item">
              <label>Response Time:</label>
              <span>{{ selectedError.response_time }}ms</span>
            </div>
            <div class="detail-item">
              <label>Error Category:</label>
              <span>{{ getErrorTypeDescription(selectedError.status_code) }}</span>
            </div>
          </div>
          <div class="error-recommendations">
            <h4>Troubleshooting Recommendations:</h4>
            <ul>
              <li v-for="rec in getErrorRecommendations(selectedError)" :key="rec">{{ rec }}</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue';
import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns';
import { useApi } from '../service/api';
import PageHeader from './PageHeader.vue';

// State
const errorData = ref([]);
const selectedError = ref(null);
const isRefreshing = ref(false);
const filters = ref({
  timeRange: '24h',
  errorType: 'all',
  statusCode: 'all',
  endpoint: 'all'
});

// Chart refs
const errorTimelineCanvas = ref(null);
const statusDistributionCanvas = ref(null);
const endpointErrorCanvas = ref(null);
const hourlyErrorCanvas = ref(null);
const responseTimeErrorCanvas = ref(null);

// Chart instances
let errorTimelineChart, statusDistributionChart, endpointErrorChart, hourlyErrorChart, responseTimeErrorChart;

const api = useApi();

// Computed properties
const uniqueStatusCodes = computed(() => {
  return [...new Set(errorData.value.filter(d => d.status_code >= 400).map(d => d.status_code))].sort();
});

const uniqueEndpoints = computed(() => {
  return [...new Set(errorData.value.filter(d => d.status_code >= 400).map(d => d.endpoint))];
});

const errorSummary = computed(() => {
  const allData = errorData.value;
  const errors = allData.filter(d => d.status_code >= 400);
  const clientErrors = errors.filter(d => d.status_code >= 400 && d.status_code < 500);
  const serverErrors = errors.filter(d => d.status_code >= 500);
  
  const statusCounts = {};
  errors.forEach(e => {
    statusCounts[e.status_code] = (statusCounts[e.status_code] || 0) + 1;
  });
  
  const mostFrequent = Object.entries(statusCounts).sort(([,a], [,b]) => b - a)[0];
  
  return {
    totalErrors: errors.length,
    errorRate: allData.length > 0 ? ((errors.length / allData.length) * 100).toFixed(1) : 0,
    clientErrors: clientErrors.length,
    clientErrorRate: errors.length > 0 ? ((clientErrors.length / errors.length) * 100).toFixed(1) : 0,
    serverErrors: serverErrors.length,
    serverErrorRate: errors.length > 0 ? ((serverErrors.length / errors.length) * 100).toFixed(1) : 0,
    mostFrequentStatus: mostFrequent ? mostFrequent[0] : 'None',
    mostFrequentCount: mostFrequent ? mostFrequent[1] : 0
  };
});

const recentErrors = computed(() => {
  return errorData.value
    .filter(d => d.status_code >= 400)
    .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
    .slice(0, 50);
});

const errorPatterns = computed(() => {
  const errors = errorData.value.filter(d => d.status_code >= 400);
  const patterns = [];
  
  // 404 Pattern
  const notFoundErrors = errors.filter(e => e.status_code === 404);
  if (notFoundErrors.length > 0) {
    patterns.push({
      id: '404-pattern',
      title: 'Resource Not Found (404)',
      count: notFoundErrors.length,
      description: 'High number of 404 errors indicating missing resources or broken links',
      recommendation: 'Review URL patterns and ensure proper routing configuration'
    });
  }
  
  // 500 Pattern
  const serverErrors = errors.filter(e => e.status_code >= 500);
  if (serverErrors.length > 0) {
    patterns.push({
      id: '500-pattern',
      title: 'Server Errors (5xx)',
      count: serverErrors.length,
      description: 'Server-side errors indicating application or infrastructure issues',
      recommendation: 'Check server logs, database connections, and application health'
    });
  }
  
  // Rate limiting pattern
  const rateLimitErrors = errors.filter(e => e.status_code === 429);
  if (rateLimitErrors.length > 0) {
    patterns.push({
      id: '429-pattern',
      title: 'Rate Limiting (429)',
      count: rateLimitErrors.length,
      description: 'Clients hitting rate limits',
      recommendation: 'Consider adjusting rate limits or implementing client-side throttling'
    });
  }
  
  return patterns;
});

const errorImpact = computed(() => {
  const errors = errorData.value.filter(d => d.status_code >= 400);
  const highImpact = errors.filter(e => e.status_code >= 500).length;
  const mediumImpact = errors.filter(e => e.status_code >= 400 && e.status_code < 500 && 
    ['POST', 'PUT', 'DELETE'].includes(e.method)).length;
  const lowImpact = errors.length - highImpact - mediumImpact;
  
  return { highImpact, mediumImpact, lowImpact };
});

// Header stats for PageHeader component
const headerStats = computed(() => {
  return [
    {
      label: 'Total Errors',
      value: errorSummary.value.totalErrors.toLocaleString(),
      type: errorSummary.value.totalErrors > 100 ? 'critical' : errorSummary.value.totalErrors > 50 ? 'warning' : 'success'
    },
    {
      label: 'Error Rate',
      value: `${errorSummary.value.errorRate}%`,
      type: errorSummary.value.errorRate > 10 ? 'critical' : errorSummary.value.errorRate > 5 ? 'warning' : 'success'
    },
    {
      label: 'Client Errors',
      value: errorSummary.value.clientErrors.toLocaleString(),
      type: 'warning'
    },
    {
      label: 'Server Errors',
      value: errorSummary.value.serverErrors.toLocaleString(),
      type: errorSummary.value.serverErrors > 0 ? 'critical' : 'success'
    }
  ];
});

// Methods
const loadErrorData = async () => {
  try {
    const params = { timeRange: filters.value.timeRange };
    if (filters.value.endpoint !== 'all') params.endpoint = filters.value.endpoint;
    
    const response = await api.getHistory(params);
    errorData.value = response.data || [];
    
    await nextTick();
    updateCharts();
  } catch (error) {
    console.error('Failed to load error data:', error);
    errorData.value = [];
  }
};

const initCharts = () => {
  Chart.defaults.color = '#d9d9d9';
  Chart.defaults.font.family = "'Roboto', sans-serif";
  
  // Error Timeline Chart
  errorTimelineChart = new Chart(errorTimelineCanvas.value, {
    type: 'line',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { type: 'time', time: { unit: 'hour' } },
        y: { beginAtZero: true, title: { display: true, text: 'Error Count' } }
      },
      plugins: {
        legend: { display: true, position: 'top' }
      }
    }
  });
  
  // Status Distribution Chart
  statusDistributionChart = new Chart(statusDistributionCanvas.value, {
    type: 'doughnut',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: { position: 'bottom' }
      }
    }
  });
  
  // Endpoint Error Chart
  endpointErrorChart = new Chart(endpointErrorCanvas.value, {
    type: 'bar',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      indexAxis: 'y',
      scales: {
        x: { beginAtZero: true, title: { display: true, text: 'Error Count' } }
      }
    }
  });
  
  // Hourly Error Chart
  hourlyErrorChart = new Chart(hourlyErrorCanvas.value, {
    type: 'bar',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { title: { display: true, text: 'Hour of Day' } },
        y: { beginAtZero: true, title: { display: true, text: 'Error Rate (%)' } }
      }
    }
  });
  
  // Response Time vs Error Chart
  responseTimeErrorChart = new Chart(responseTimeErrorCanvas.value, {
    type: 'scatter',
    data: { datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { title: { display: true, text: 'Response Time (ms)' } },
        y: { title: { display: true, text: 'Error Rate (%)' } }
      }
    }
  });
};

const updateCharts = () => {
  const errors = errorData.value.filter(d => d.status_code >= 400);
  const allData = errorData.value;
  
  // Error Timeline
  const hourlyData = {};
  errors.forEach(e => {
    const hour = new Date(e.timestamp).toISOString().slice(0, 13) + ':00:00.000Z';
    if (!hourlyData[hour]) hourlyData[hour] = { client: 0, server: 0 };
    if (e.status_code >= 400 && e.status_code < 500) hourlyData[hour].client++;
    else if (e.status_code >= 500) hourlyData[hour].server++;
  });
  
  const labels = Object.keys(hourlyData).sort();
  errorTimelineChart.data.labels = labels;
  errorTimelineChart.data.datasets = [
    {
      label: '4xx Client Errors',
      data: labels.map(l => hourlyData[l]?.client || 0),
      borderColor: '#ffb74d',
      backgroundColor: 'rgba(255, 183, 77, 0.1)',
      tension: 0.4
    },
    {
      label: '5xx Server Errors',
      data: labels.map(l => hourlyData[l]?.server || 0),
      borderColor: '#f2495c',
      backgroundColor: 'rgba(242, 73, 92, 0.1)',
      tension: 0.4
    }
  ];
  errorTimelineChart.update();
  
  // Status Distribution
  const statusCounts = {};
  errors.forEach(e => {
    statusCounts[e.status_code] = (statusCounts[e.status_code] || 0) + 1;
  });
  
  statusDistributionChart.data.labels = Object.keys(statusCounts);
  statusDistributionChart.data.datasets = [{
    data: Object.values(statusCounts),
    backgroundColor: ['#f2495c', '#ffb74d', '#b877d9', '#fade2a', '#73bf69']
  }];
  statusDistributionChart.update();
  
  // Endpoint Errors
  const endpointCounts = {};
  errors.forEach(e => {
    endpointCounts[e.endpoint] = (endpointCounts[e.endpoint] || 0) + 1;
  });
  
  const topEndpoints = Object.entries(endpointCounts)
    .sort(([,a], [,b]) => b - a)
    .slice(0, 10);
  
  endpointErrorChart.data.labels = topEndpoints.map(([ep]) => ep);
  endpointErrorChart.data.datasets = [{
    label: 'Error Count',
    data: topEndpoints.map(([,count]) => count),
    backgroundColor: '#f2495c'
  }];
  endpointErrorChart.update();
  
  // Hourly Error Rate
  const hourlyStats = new Array(24).fill(0).map(() => ({ total: 0, errors: 0 }));
  allData.forEach(d => {
    const hour = new Date(d.timestamp).getHours();
    hourlyStats[hour].total++;
    if (d.status_code >= 400) hourlyStats[hour].errors++;
  });
  
  hourlyErrorChart.data.labels = Array.from({length: 24}, (_, i) => `${i}:00`);
  hourlyErrorChart.data.datasets = [{
    label: 'Error Rate (%)',
    data: hourlyStats.map(h => h.total > 0 ? (h.errors / h.total * 100) : 0),
    backgroundColor: '#f2495c'
  }];
  hourlyErrorChart.update();
  
  // Response Time vs Error correlation would need more complex analysis
  responseTimeErrorChart.update();
};

// Utility methods
const formatTimestamp = (timestamp) => {
  return new Date(timestamp).toLocaleString();
};

const getErrorRowClass = (statusCode) => {
  if (statusCode >= 500) return 'error-row server-error';
  if (statusCode >= 400) return 'error-row client-error';
  return 'error-row';
};

const getStatusClass = (statusCode) => {
  if (statusCode >= 500) return 'status-badge server';
  if (statusCode >= 400) return 'status-badge client';
  return 'status-badge';
};

const getMethodClass = (method) => {
  const methodColors = {
    'GET': 'method-badge get',
    'POST': 'method-badge post',
    'PUT': 'method-badge put',
    'DELETE': 'method-badge delete',
    'PATCH': 'method-badge patch'
  };
  return methodColors[method] || 'method-badge';
};

const getErrorTypeDescription = (statusCode) => {
  if (statusCode >= 500) return 'Server Error';
  if (statusCode >= 400) return 'Client Error';
  return 'Unknown';
};

const showErrorDetails = (error) => {
  selectedError.value = error;
};

const closeErrorDetails = () => {
  selectedError.value = null;
};

const copyErrorInfo = (error) => {
  const info = `Status: ${error.status_code}\nEndpoint: ${error.method} ${error.endpoint}\nTime: ${formatTimestamp(error.timestamp)}\nResponse Time: ${error.response_time}ms`;
  navigator.clipboard.writeText(info);
};

const getErrorRecommendations = (error) => {
  const recommendations = [];
  
  if (error.status_code === 404) {
    recommendations.push('Check if the requested resource exists');
    recommendations.push('Verify URL routing configuration');
    recommendations.push('Ensure proper API documentation');
  } else if (error.status_code >= 500) {
    recommendations.push('Check server logs for detailed error messages');
    recommendations.push('Verify database connections and health');
    recommendations.push('Monitor application memory and CPU usage');
    recommendations.push('Check for any recent deployments or changes');
  } else if (error.status_code === 429) {
    recommendations.push('Implement client-side request throttling');
    recommendations.push('Review rate limiting policies');
    recommendations.push('Consider using exponential backoff');
  } else if (error.status_code === 401 || error.status_code === 403) {
    recommendations.push('Verify authentication credentials');
    recommendations.push('Check user permissions and roles');
    recommendations.push('Review API key or token validity');
  }
  
  return recommendations;
};

// Handler functions
const handleRefresh = async () => {
  isRefreshing.value = true;
  await loadErrorData();
  isRefreshing.value = false;
};

const handleExport = () => {
  const data = {
    timestamp: new Date().toISOString(),
    filters: filters.value,
    summary: errorSummary.value,
    patterns: errorPatterns.value,
    recentErrors: recentErrors.value.slice(0, 100),
    data: errorData.value.filter(d => d.status_code >= 400).slice(0, 1000)
  };
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `error-analysis-${new Date().toISOString().split('T')[0]}.json`;
  a.click();
  URL.revokeObjectURL(url);
};

// Lifecycle
onMounted(() => {
  initCharts();
  loadErrorData();
  setInterval(loadErrorData, 300000); // Refresh every 5 minutes
});
</script>

<style scoped>
.error-analysis-section {
  padding: 20px 0;
}

/* Filters */
.error-filters {
  display: flex;
  gap: 20px;
  margin-bottom: 30px;
  padding: 15px;
  background-color: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.filter-label {
  font-size: 12px;
  color: var(--grafana-text-muted);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.filter-select {
  padding: 8px 12px;
  background-color: var(--grafana-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
  color: var(--grafana-text);
  font-size: 14px;
  min-width: 150px;
}

.refresh-button {
  padding: 8px 16px;
  background-color: var(--grafana-blue);
  border: none;
  border-radius: 2px;
  color: #000;
  font-size: 14px;
  cursor: pointer;
  align-self: flex-end;
}

/* Summary Cards */
.error-summary {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 30px;
}

.summary-card {
  background-color: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
  padding: 20px;
  position: relative;
}

.summary-card.error {
  border-left: 4px solid var(--grafana-red);
}

.summary-card.client {
  border-left: 4px solid var(--grafana-orange);
}

.summary-card.server {
  border-left: 4px solid var(--grafana-purple);
}

.summary-card.frequent {
  border-left: 4px solid var(--grafana-blue);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 14px;
  color: var(--grafana-text-muted);
}

.card-value {
  font-size: 32px;
  font-weight: 700;
  color: var(--grafana-text);
  margin-bottom: 5px;
}

.card-subtitle {
  font-size: 12px;
  color: var(--grafana-text-muted);
}

/* Grid */
.error-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 20px;
}

.panel {
  background-color: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
}

.panel.col-12 { grid-column: span 12; }
.panel.col-6 { grid-column: span 6; }

.panel-header {
  padding: 15px 20px;
  border-bottom: 1px solid var(--grafana-border);
}

.panel-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--grafana-text);
  margin-bottom: 4px;
}

.panel-subtitle {
  font-size: 12px;
  color: var(--grafana-text-muted);
}

.panel-content {
  padding: 20px;
}

/* Charts */
.chart-container.large-chart {
  height: 350px;
}

.chart-container.medium-chart {
  height: 300px;
}

/* Error Table */
.error-table-container {
  max-height: 500px;
  overflow-y: auto;
}

.error-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.error-table th {
  background-color: var(--grafana-bg-light);
  padding: 12px;
  text-align: left;
  font-weight: 600;
  border-bottom: 1px solid var(--grafana-border);
  position: sticky;
  top: 0;
  z-index: 1;
}

.error-table td {
  padding: 12px;
  border-bottom: 1px solid var(--grafana-border);
}

.error-row.server-error {
  background-color: rgba(242, 73, 92, 0.05);
}

.error-row.client-error {
  background-color: rgba(255, 183, 77, 0.05);
}

.status-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.server {
  background-color: var(--grafana-red);
  color: white;
}

.status-badge.client {
  background-color: var(--grafana-orange);
  color: white;
}

.method-badge {
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.method-badge.get { background-color: var(--grafana-blue); color: white; }
.method-badge.post { background-color: var(--grafana-green); color: white; }
.method-badge.put { background-color: var(--grafana-orange); color: white; }
.method-badge.delete { background-color: var(--grafana-red); color: white; }
.method-badge.patch { background-color: var(--grafana-purple); color: white; }

.action-btn {
  background: none;
  border: none;
  color: var(--grafana-text-muted);
  cursor: pointer;
  padding: 4px;
  margin: 0 2px;
}

.action-btn:hover {
  color: var(--grafana-blue);
}

.no-errors {
  text-align: center;
  padding: 40px;
  color: var(--grafana-text-muted);
}

.no-errors i {
  font-size: 48px;
  color: var(--grafana-green);
  margin-bottom: 15px;
}

/* Patterns */
.pattern-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.pattern-item {
  padding: 15px;
  background-color: var(--grafana-bg-light);
  border-radius: 4px;
  border-left: 4px solid var(--grafana-orange);
}

.pattern-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.pattern-title {
  font-weight: 600;
  color: var(--grafana-text);
}

.pattern-count {
  background-color: var(--grafana-orange);
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 12px;
}

.pattern-description {
  font-size: 14px;
  color: var(--grafana-text-muted);
  margin-bottom: 8px;
}

.pattern-recommendation {
  font-size: 13px;
  color: var(--grafana-blue);
  display: flex;
  align-items: center;
  gap: 6px;
}

/* Impact Analysis */
.impact-metrics {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.impact-item {
  padding: 15px;
  border-radius: 4px;
  text-align: center;
}

.impact-item.high {
  background-color: rgba(242, 73, 92, 0.1);
  border: 1px solid var(--grafana-red);
}

.impact-item.medium {
  background-color: rgba(255, 183, 77, 0.1);
  border: 1px solid var(--grafana-orange);
}

.impact-item.low {
  background-color: rgba(115, 191, 105, 0.1);
  border: 1px solid var(--grafana-green);
}

.impact-label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
  color: var(--grafana-text-muted);
}

.impact-value {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 5px;
}

.impact-desc {
  font-size: 12px;
  color: var(--grafana-text-muted);
}

/* Modal */
.error-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.error-modal {
  background-color: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 4px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid var(--grafana-border);
}

.modal-header h3 {
  margin: 0;
  color: var(--grafana-text);
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--grafana-text-muted);
  cursor: pointer;
}

.modal-content {
  padding: 20px;
}

.error-detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
  margin-bottom: 25px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.detail-item label {
  font-size: 12px;
  color: var(--grafana-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-item span {
  font-size: 14px;
  color: var(--grafana-text);
  font-weight: 500;
}

.error-recommendations h4 {
  color: var(--grafana-text);
  margin-bottom: 15px;
}

.error-recommendations ul {
  list-style: none;
  padding: 0;
}

.error-recommendations li {
  padding: 8px 0;
  color: var(--grafana-text-muted);
  border-bottom: 1px solid var(--grafana-border);
}

.error-recommendations li:before {
  content: "→ ";
  color: var(--grafana-blue);
  font-weight: 600;
}

/* Responsive */
@media (max-width: 1200px) {
  .error-summary {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .panel.col-6 {
    grid-column: span 12;
  }
}

@media (max-width: 768px) {
  .error-filters {
    flex-direction: column;
    gap: 15px;
  }
  
  .error-summary {
    grid-template-columns: 1fr;
  }
  
  .error-grid {
    grid-template-columns: 1fr;
  }
  
  .panel.col-12,
  .panel.col-6 {
    grid-column: span 1;
  }
  
  .error-detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
