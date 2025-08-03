<template>
  <div class="timeline-section">
    <!-- Page Header -->
    <PageHeader
      page-title="Request Timeline Analysis"
      page-description="Detailed timeline view of API requests with performance metrics and patterns"
      page-icon="fas fa-chart-area"
      :stats="headerStats"
      :refreshing="isRefreshing"
      :show-export="true"
      @refresh="handleRefresh"
      @export="handleExport"
    >
      <template #filters>
        <div class="timeline-filter-group">
          <label class="timeline-filter-label">Time Range</label>
          <select id="timeline-time-range" class="analytics-select" v-model="filters.timeRange" @change="loadTimelineData">
            <option value="1h">Last 1 Hour</option>
            <option value="6h">Last 6 Hours</option>
            <option value="24h">Last 24 Hours</option>
            <option value="168h">Last 7 Days</option>
          </select>
        </div>
        <div class="timeline-filter-group">
          <label class="timeline-filter-label">Endpoint Filter</label>
          <select id="timeline-endpoint-filter" class="analytics-select" v-model="filters.endpoint" @change="loadTimelineData">
            <option value="all">All Endpoints</option>
            <option v-for="ep in uniqueEndpoints" :key="ep" :value="ep">{{ ep }}</option>
          </select>
        </div>
        <div class="timeline-filter-group">
          <label class="timeline-filter-label">Method Filter</label>
          <select id="timeline-method-filter" class="analytics-select" v-model="filters.method" @change="loadTimelineData">
            <option value="all">All Methods</option>
            <option value="GET">GET</option>
            <option value="POST">POST</option>
            <option value="PUT">PUT</option>
            <option value="DELETE">DELETE</option>
            <option value="PATCH">PATCH</option>
          </select>
        </div>
      </template>
    </PageHeader>

    <!-- Timeline Stats -->
    <div class="timeline-stats">
      <div class="timeline-stat-card">
        <div class="timeline-stat-value">{{ timelineMetrics.totalRequests }}</div>
        <div class="timeline-stat-label">Total Requests</div>
      </div>
      <div class="timeline-stat-card">
        <div class="timeline-stat-value">{{ timelineMetrics.uniqueEndpoints }}</div>
        <div class="timeline-stat-label">Unique Endpoints</div>
      </div>
      <div class="timeline-stat-card">
        <div class="timeline-stat-value">{{ timelineMetrics.errorRate }}%</div>
        <div class="timeline-stat-label">Error Rate</div>
      </div>
      <div class="timeline-stat-card">
        <div class="timeline-stat-value">{{ timelineMetrics.avgResponse }}ms</div>
        <div class="timeline-stat-label">Avg Response Time</div>
      </div>
    </div>

    <!-- Timeline Charts Container -->
    <div id="timeline-charts-container">
      <div v-for="endpoint in groupedEndpoints" :key="endpoint.name" class="timeline-endpoint-chart">
        <div class="timeline-endpoint-title">{{ endpoint.name }}</div>
        <div class="timeline-chart-container">
          <canvas :ref="'chart-' + endpoint.id" :id="'chart-' + endpoint.id"></canvas>
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

const historyData = ref([]);
const isRefreshing = ref(false);
const filters = ref({
  timeRange: '24h',
  endpoint: 'all',
  method: 'all'
});
const timelineCharts = ref({});
const api = useApi();

// Computed properties
const uniqueEndpoints = computed(() => {
  return [...new Set(historyData.value.map(d => d.endpoint))];
});

const timelineMetrics = computed(() => {
  const data = historyData.value;
  if (!data.length) return { totalRequests: 0, uniqueEndpoints: 0, errorRate: 0, avgResponse: 0 };
  
  const totalRequests = data.length;
  const uniqueEndpoints = new Set(data.map(d => d.endpoint)).size;
  const errors = data.filter(d => d.status_code >= 400);
  const errorRate = ((errors.length / totalRequests) * 100).toFixed(1);
  const avgResponse = (data.reduce((sum, d) => sum + d.response_time, 0) / totalRequests).toFixed(0);
  
  return { totalRequests, uniqueEndpoints, errorRate, avgResponse };
});

const groupedEndpoints = computed(() => {
  const groups = {};
  historyData.value.forEach(d => {
    const key = `${d.method} ${d.endpoint}`;
    if (!groups[key]) {
      groups[key] = {
        name: key,
        id: key.replace(/\s+/g, '-').replace(/[^a-zA-Z0-9-]/g, ''),
        data: []
      };
    }
    groups[key].data.push(d);
  });
  return Object.values(groups);
});

// Header stats for PageHeader component
const headerStats = computed(() => {
  return [
    {
      label: 'Total Requests',
      value: timelineMetrics.value.totalRequests.toLocaleString(),
      type: 'info'
    },
    {
      label: 'Unique Endpoints',
      value: timelineMetrics.value.uniqueEndpoints.toString(),
      type: 'info'
    },
    {
      label: 'Error Rate',
      value: `${timelineMetrics.value.errorRate}%`,
      type: timelineMetrics.value.errorRate > 10 ? 'critical' : timelineMetrics.value.errorRate > 5 ? 'warning' : 'success'
    },
    {
      label: 'Avg Response',
      value: `${timelineMetrics.value.avgResponse}ms`,
      type: timelineMetrics.value.avgResponse > 500 ? 'critical' : timelineMetrics.value.avgResponse > 200 ? 'warning' : 'success'
    }
  ];
});

// Load timeline data from history API
const loadTimelineData = async () => {
  try {
    const params = {
      timeRange: filters.value.timeRange
    };
    if (filters.value.endpoint !== 'all') params.endpoint = filters.value.endpoint;
    if (filters.value.method !== 'all') params.method = filters.value.method;
    
    const response = await api.getHistory(params);
    historyData.value = response.data || [];
    
    await nextTick();
    updateTimelineCharts();
  } catch (error) {
    console.error('Failed to load timeline data:', error);
    historyData.value = [];
  }
};

// Update timeline charts
const updateTimelineCharts = () => {
  groupedEndpoints.value.forEach(endpoint => {
    const canvasId = `chart-${endpoint.id}`;
    const canvas = document.getElementById(canvasId);
    if (!canvas) return;
    
    // Destroy existing chart
    if (timelineCharts.value[endpoint.id]) {
      timelineCharts.value[endpoint.id].destroy();
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
    
    // Create complete minute buckets with zero values
    const timeLabels = [];
    const status2xxMap = {};
    const status4xxMap = {};
    const status5xxMap = {};
    
    for (let i = minutes - 1; i >= 0; i--) {
      const time = new Date(now.getTime() - i * 60 * 1000); // minutes ago
      const timeKey = time.toISOString().slice(0, 16) + ':00.000Z';
      timeLabels.push(timeKey);
      status2xxMap[timeKey] = 0;
      status4xxMap[timeKey] = 0;
      status5xxMap[timeKey] = 0;
    }
    
    // Fill with actual data
    endpoint.data.forEach(d => {
      const minute = new Date(d.timestamp || Date.now()).toISOString().slice(0, 16) + ':00.000Z';
      
      if (status2xxMap.hasOwnProperty(minute)) {
        if (d.status_code >= 200 && d.status_code < 300) {
          status2xxMap[minute]++;
        } else if (d.status_code >= 400 && d.status_code < 500) {
          status4xxMap[minute]++;
        } else if (d.status_code >= 500) {
          status5xxMap[minute]++;
        }
      }
    });
    
    // Create chart
    const ctx = canvas.getContext('2d');
    timelineCharts.value[endpoint.id] = new Chart(ctx, {
      type: 'line',
      data: {
        labels: timeLabels,
        datasets: [
          {
            label: '2xx Success',
            data: timeLabels.map(l => status2xxMap[l]),
            borderColor: '#73bf69',
            backgroundColor: 'rgba(115, 191, 105, 0.1)',
            fill: true,
            tension: 0.3,
            spanGaps: false,
            pointRadius: 1
          },
          {
            label: '4xx Client Error',
            data: timeLabels.map(l => status4xxMap[l]),
            borderColor: '#ffb74d',
            backgroundColor: 'rgba(255, 183, 77, 0.1)',
            fill: true,
            tension: 0.3,
            spanGaps: false,
            pointRadius: 1
          },
          {
            label: '5xx Server Error',
            data: timeLabels.map(l => status5xxMap[l]),
            borderColor: '#f2495c',
            backgroundColor: 'rgba(242, 73, 92, 0.1)',
            fill: true,
            tension: 0.3,
            spanGaps: false,
            pointRadius: 1
          }
        ]
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
          legend: { position: 'top' },
          tooltip: {
            mode: 'nearest',
            intersect: false
          }
        },
        interaction: {
          mode: 'nearest',
          axis: 'x',
          intersect: false
        }
      }
    });
  });
};

// Handler functions
const handleRefresh = async () => {
  isRefreshing.value = true;
  await loadTimelineData();
  isRefreshing.value = false;
};

const handleExport = () => {
  const data = {
    timestamp: new Date().toISOString(),
    filters: filters.value,
    metrics: timelineMetrics.value,
    groupedData: groupedEndpoints.value,
    data: historyData.value
  };
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `timeline-analysis-${new Date().toISOString().split('T')[0]}.json`;
  a.click();
  URL.revokeObjectURL(url);
};

onMounted(() => {
  loadTimelineData();
});
</script>

<style scoped>
.timeline-filters {
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

.timeline-filter-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.timeline-filter-label {
  font-size: 12px;
  color: var(--grafana-text-muted);
  font-weight: 500;
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

.timeline-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.timeline-stat-card {
  background: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 4px;
  padding: 15px;
  text-align: center;
}

.timeline-stat-value {
  font-size: 24px;
  font-weight: 500;
  color: var(--grafana-blue);
  margin-bottom: 5px;
}

.timeline-stat-label {
  font-size: 12px;
  color: var(--grafana-text-muted);
}

#timeline-charts-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
  gap: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.timeline-endpoint-chart {
  background: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 4px;
  padding: 15px;
  max-width: 600px;
}

.timeline-endpoint-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--grafana-text);
  margin-bottom: 15px;
  border-bottom: 1px solid var(--grafana-border);
  padding-bottom: 10px;
}

.timeline-chart-container {
  height: 200px;
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
}
</style>
