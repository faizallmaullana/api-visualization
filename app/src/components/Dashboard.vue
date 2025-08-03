<template>
  <div class="analytics-section">
    <!-- Analytics Filters -->
    <div class="analytics-filters">
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Time Range</label>
        <select class="analytics-select" v-model="filters.timeRange" @change="loadAnalyticsData">
          <option value="1h">Last 1 Hour</option>
          <option value="6h">Last 6 Hours</option>
          <option value="24h">Last 24 Hours</option>
          <option value="168h">Last 7 Days</option>
        </select>
      </div>
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Endpoint Filter</label>
        <select class="analytics-select" v-model="filters.endpoint" @change="loadAnalyticsData">
          <option value="all">All Endpoints</option>
          <option v-for="ep in uniqueEndpoints" :key="ep" :value="ep">{{ ep }}</option>
        </select>
      </div>
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Method Filter</label>
        <select class="analytics-select" v-model="filters.method" @change="loadAnalyticsData">
          <option value="all">All Methods</option>
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="DELETE">DELETE</option>
          <option value="PATCH">PATCH</option>
        </select>
      </div>
      <button class="analytics-button" @click="loadAnalyticsData">
        <i class="fas fa-sync-alt"></i> Refresh
      </button>
    </div>

    <!-- Key Metrics Grid -->
    <div class="dashboard-grid">
      <!-- Response Time Metrics -->
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">Average Response Time</div></div>
        <div class="panel-content">
          <div class="stat-value stat-blue">{{ responseMetrics.avgResponse }}ms</div>
          <div class="stat-unit">milliseconds</div>
        </div>
      </div>
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">P95 Response Time</div></div>
        <div class="panel-content">
          <div class="stat-value stat-red">{{ responseMetrics.p95Response }}ms</div>
          <div class="stat-unit">95th percentile</div>
        </div>
      </div>

      <!-- RPS Metrics -->
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">Current RPS</div></div>
        <div class="panel-content">
          <div class="stat-value stat-blue">{{ rpsMetrics.currentRPS }}</div>
          <div class="stat-unit">requests/sec</div>
        </div>
      </div>
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">Peak RPS</div></div>
        <div class="panel-content">
          <div class="stat-value stat-orange">{{ rpsMetrics.peakRPS }}</div>
          <div class="stat-unit">highest</div>
        </div>
      </div>

      <!-- Quality Metrics -->
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">Error Rate</div></div>
        <div class="panel-content">
          <div class="stat-value stat-red">{{ qualityMetrics.errorRate }}%</div>
          <div class="stat-unit">4xx/5xx errors</div>
        </div>
      </div>
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">Success Rate</div></div>
        <div class="panel-content">
          <div class="stat-value stat-green">{{ qualityMetrics.successRate }}%</div>
          <div class="stat-unit">2xx responses</div>
        </div>
      </div>
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">Total Requests</div></div>
        <div class="panel-content">
          <div class="stat-value stat-purple">{{ rpsMetrics.totalRequests }}</div>
          <div class="stat-unit">total</div>
        </div>
      </div>
      <div class="panel stat-panel col-3">
        <div class="panel-header"><div class="panel-title">Active Endpoints</div></div>
        <div class="panel-content">
          <div class="stat-value stat-orange">{{ throughputMetrics.activeEndpoints }}</div>
          <div class="stat-unit">routes</div>
        </div>
      </div>

      <!-- Charts -->
      <div class="panel col-12">
        <div class="panel-header">
          <div class="panel-title">Requests Per Second Timeline</div>
        </div>
        <div class="panel-content">
          <div class="chart-container time-series-chart">
            <canvas ref="rpsTimelineCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">Response Time Distribution</div>
        </div>
        <div class="panel-content">
          <div class="chart-container small-chart">
            <canvas ref="distributionCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel col-6">
        <div class="panel-header">
          <div class="panel-title">HTTP Status Code Distribution</div>
        </div>
        <div class="panel-content">
          <div class="chart-container small-chart">
            <canvas ref="statusDistributionCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel col-8">
        <div class="panel-header">
          <div class="panel-title">Response Time by Endpoint</div>
        </div>
        <div class="panel-content">
          <div class="chart-container medium-chart">
            <canvas ref="endpointCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel col-4">
        <div class="panel-header">
          <div class="panel-title">API Health Score</div>
        </div>
        <div class="panel-content">
          <div class="health-score-container">
            <div class="health-score-value" :class="healthScoreClass">{{ apiHealthScore }}</div>
            <div class="health-score-label">Overall Health</div>
            <div class="health-indicators">
              <div class="health-indicator" :class="{ good: qualityMetrics.successRate >= 95 }">
                <span>Success: {{ qualityMetrics.successRate }}%</span>
              </div>
              <div class="health-indicator" :class="{ good: responseMetrics.avgResponse <= 500 }">
                <span>Avg RT: {{ responseMetrics.avgResponse }}ms</span>
              </div>
              <div class="health-indicator" :class="{ good: qualityMetrics.errorRate <= 5 }">
                <span>Errors: {{ qualityMetrics.errorRate }}%</span>
              </div>
            </div>
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

const historyData = ref([]);
const filters = ref({
  timeRange: '24h',
  endpoint: 'all',
  method: 'all'
});

const rpsTimelineCanvas = ref(null);
const distributionCanvas = ref(null);
const statusDistributionCanvas = ref(null);
const endpointCanvas = ref(null);

let rpsTimelineChart, distributionChart, statusDistributionChart, endpointChart;
const api = useApi();

// Computed properties
const uniqueEndpoints = computed(() => {
  return [...new Set(historyData.value.map(d => d.endpoint))];
});

const responseMetrics = computed(() => {
  const data = historyData.value;
  if (!data.length) return { avgResponse: 0, p95Response: 0 };
  
  const responseTimes = data.map(d => d.response_time).sort((a, b) => a - b);
  const avgResponse = (responseTimes.reduce((sum, t) => sum + t, 0) / responseTimes.length).toFixed(1);
  const p95Index = Math.floor(responseTimes.length * 0.95);
  const p95Response = responseTimes[p95Index]?.toFixed(1) || 0;
  
  return { avgResponse, p95Response };
});

const rpsMetrics = computed(() => {
  const data = historyData.value;
  if (!data.length) return { currentRPS: 0, peakRPS: 0, totalRequests: 0 };
  
  const timeRangeMap = { '1h': 3600, '6h': 21600, '24h': 86400, '168h': 604800 };
  const timeRangeSeconds = timeRangeMap[filters.value.timeRange] || 3600;
  
  const requestsBySecond = {};
  data.forEach(d => {
    const timestamp = new Date(d.timestamp || Date.now());
    const secondKey = Math.floor(timestamp.getTime() / 1000);
    requestsBySecond[secondKey] = (requestsBySecond[secondKey] || 0) + 1;
  });
  
  const rpsValues = Object.values(requestsBySecond);
  const totalRequests = data.length;
  const peakRPS = rpsValues.length ? Math.max(...rpsValues) : 0;
  
  const now = Math.floor(Date.now() / 1000);
  const lastMinuteRequests = Object.entries(requestsBySecond)
    .filter(([sec]) => now - parseInt(sec) <= 60)
    .reduce((sum, [, count]) => sum + count, 0);
  const currentRPS = (lastMinuteRequests / 60).toFixed(1);
  
  return { currentRPS, peakRPS, totalRequests };
});

const qualityMetrics = computed(() => {
  const data = historyData.value;
  if (!data.length) return { errorRate: 0, successRate: 100 };
  
  const totalRequests = data.length;
  const successRequests = data.filter(d => d.status_code >= 200 && d.status_code < 400).length;
  const errorRequests = data.filter(d => d.status_code >= 400).length;
  
  const errorRate = ((errorRequests / totalRequests) * 100).toFixed(1);
  const successRate = ((successRequests / totalRequests) * 100).toFixed(1);
  
  return { errorRate, successRate };
});

const throughputMetrics = computed(() => {
  const data = historyData.value;
  const activeEndpoints = new Set(data.map(d => `${d.method} ${d.endpoint}`)).size;
  return { activeEndpoints };
});

const apiHealthScore = computed(() => {
  const { successRate, errorRate } = qualityMetrics.value;
  const { avgResponse } = responseMetrics.value;
  
  let score = 100;
  score -= (100 - parseFloat(successRate)) * 0.5;
  score -= parseFloat(errorRate) * 0.3;
  score -= Math.max(0, (parseFloat(avgResponse) - 200) / 10);
  
  return Math.max(0, Math.min(100, score)).toFixed(0);
});

const healthScoreClass = computed(() => {
  const score = parseFloat(apiHealthScore.value);
  if (score >= 90) return 'excellent';
  if (score >= 70) return 'good';
  if (score >= 50) return 'warning';
  return 'critical';
});

// Load analytics data
const loadAnalyticsData = async () => {
  try {
    const params = {
      timeRange: filters.value.timeRange
    };
    if (filters.value.endpoint !== 'all') params.endpoint = filters.value.endpoint;
    if (filters.value.method !== 'all') params.method = filters.value.method;
    
    const response = await api.getHistory(params);
    historyData.value = response.data || [];
    
    await nextTick();
    updateCharts();
  } catch (error) {
    console.error('Failed to load analytics data:', error);
    historyData.value = [];
  }
};

// Initialize charts
const initCharts = () => {
  Chart.defaults.color = '#d9d9d9';
  Chart.defaults.font.family = "'Roboto', sans-serif";
  
  rpsTimelineChart = new Chart(rpsTimelineCanvas.value, {
    type: 'line',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { type: 'time', time: { unit: 'minute' } },
        y: { beginAtZero: true, title: { display: true, text: 'Requests Per Second' } }
      }
    }
  });

  distributionChart = new Chart(distributionCanvas.value, {
    type: 'bar',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { title: { display: true, text: 'Response Time Range (ms)' } },
        y: { beginAtZero: true, title: { display: true, text: 'Request Count' } }
      }
    }
  });

  statusDistributionChart = new Chart(statusDistributionCanvas.value, {
    type: 'doughnut',
    data: { labels: [], datasets: [] },
    options: { responsive: true, maintainAspectRatio: false }
  });

  endpointChart = new Chart(endpointCanvas.value, {
    type: 'bar',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      indexAxis: 'y',
      scales: {
        x: { beginAtZero: true, title: { display: true, text: 'Average Response Time (ms)' } }
      }
    }
  });
};

// Update charts
const updateCharts = () => {
  const data = historyData.value;
  
  // RPS Timeline
  const minutelyRPS = {};
  data.forEach(d => {
    const minute = new Date(d.timestamp || Date.now()).toISOString().slice(0, 16) + ':00.000Z';
    minutelyRPS[minute] = (minutelyRPS[minute] || 0) + 1;
  });

  const rpsLabels = Object.keys(minutelyRPS).sort();
  const rpsValues = rpsLabels.map(label => (minutelyRPS[label] / 60).toFixed(2));

  rpsTimelineChart.data.labels = rpsLabels;
  rpsTimelineChart.data.datasets = [{
    label: 'RPS',
    data: rpsValues,
    borderColor: '#52c5f7',
    backgroundColor: 'rgba(82, 197, 247, 0.1)',
    fill: true,
    tension: 0.4
  }];
  rpsTimelineChart.update();

  // Response Time Distribution
  const buckets = [0, 100, 250, 500, 1000, 2000, 5000];
  const bucketCounts = new Array(buckets.length - 1).fill(0);
  const bucketLabels = buckets.slice(0, -1).map((b, i) => `${b}-${buckets[i + 1]}ms`);

  data.forEach(d => {
    for (let i = 0; i < buckets.length - 1; i++) {
      if (d.response_time >= buckets[i] && d.response_time < buckets[i + 1]) {
        bucketCounts[i]++;
        break;
      }
    }
  });

  distributionChart.data.labels = bucketLabels;
  distributionChart.data.datasets = [{
    label: 'Request Count',
    data: bucketCounts,
    backgroundColor: 'rgba(115, 191, 105, 0.8)',
    borderColor: '#73bf69',
    borderWidth: 1
  }];
  distributionChart.update();

  // Status Distribution
  const statusCounts = {};
  data.forEach(d => {
    const statusGroup = Math.floor(d.status_code / 100) + 'xx';
    statusCounts[statusGroup] = (statusCounts[statusGroup] || 0) + 1;
  });

  statusDistributionChart.data.labels = Object.keys(statusCounts);
  statusDistributionChart.data.datasets = [{
    data: Object.values(statusCounts),
    backgroundColor: ['#73bf69', '#fade2a', '#ffb74d', '#f2495c', '#b877d9']
  }];
  statusDistributionChart.update();

  // Endpoint Chart
  const endpointStats = {};
  data.forEach(d => {
    const key = `${d.method} ${d.endpoint}`;
    if (!endpointStats[key]) {
      endpointStats[key] = { count: 0, totalResponseTime: 0 };
    }
    endpointStats[key].count++;
    endpointStats[key].totalResponseTime += d.response_time;
  });

  const topEndpoints = Object.entries(endpointStats)
    .map(([endpoint, stats]) => ({
      endpoint,
      avgResponse: (stats.totalResponseTime / stats.count).toFixed(1)
    }))
    .sort((a, b) => parseFloat(b.avgResponse) - parseFloat(a.avgResponse))
    .slice(0, 10);

  endpointChart.data.labels = topEndpoints.map(e => e.endpoint);
  endpointChart.data.datasets = [{
    label: 'Avg Response Time (ms)',
    data: topEndpoints.map(e => parseFloat(e.avgResponse)),
    backgroundColor: 'rgba(255, 183, 77, 0.8)',
    borderColor: '#ffb74d',
    borderWidth: 1
  }];
  endpointChart.update();
};

onMounted(() => {
  initCharts();
  loadAnalyticsData();
  setInterval(loadAnalyticsData, 60000);
});
</script>

<style scoped>
.analytics-section {
  padding: 20px 0;
}

.analytics-filters {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
  padding: 15px;
  background-color: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
  align-items: end;
}

.analytics-filter-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.analytics-filter-label {
  font-size: 12px;
  color: var(--grafana-text-muted);
  font-weight: 500;
}

.analytics-select {
  padding: 8px 12px;
  background-color: var(--grafana-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
  color: var(--grafana-text);
  font-size: 14px;
}

.analytics-button {
  padding: 8px 16px;
  background-color: var(--grafana-blue);
  border: none;
  border-radius: 2px;
  color: #000;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.analytics-button:hover {
  background-color: rgba(82, 197, 247, 0.8);
}
</style>
