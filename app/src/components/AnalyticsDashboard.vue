<template>
  <div class="detailed-analytics-section">
    <!-- Filters -->
    <div class="analytics-filters">
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Time Range</label>
        <select class="analytics-select" v-model="filters.timeRange" @change="loadDetailedData">
          <option value="1h">Last 1 Hour</option>
          <option value="6h">Last 6 Hours</option>
          <option value="24h">Last 24 Hours</option>
          <option value="168h">Last 7 Days</option>
        </select>
      </div>
      <div class="analytics-filter-group">
        <label class="analytics-filter-label">Endpoint Filter</label>
        <select class="analytics-select" v-model="filters.endpoint" @change="loadDetailedData">
          <option value="all">All Endpoints</option>
          <option v-for="ep in uniqueEndpoints" :key="ep" :value="ep">{{ ep }}</option>
        </select>
      </div>
      <button class="analytics-button" @click="loadDetailedData">
        <i class="fas fa-sync-alt"></i> Refresh
      </button>
    </div>

    <!-- Main Grid -->
    <div class="dashboard-grid">
      <!-- Performance Metrics -->
      <div class="panel stat-panel col-2"><div class="panel-header"><div class="panel-title">Avg Response</div></div><div class="panel-content"><div class="stat-value stat-blue">{{ metrics.avgResponse }}ms</div></div></div>
      <div class="panel stat-panel col-2"><div class="panel-header"><div class="panel-title">P95 Response</div></div><div class="panel-content"><div class="stat-value stat-orange">{{ metrics.p95Response }}ms</div></div></div>
      <div class="panel stat-panel col-2"><div class="panel-header"><div class="panel-title">P99 Response</div></div><div class="panel-content"><div class="stat-value stat-red">{{ metrics.p99Response }}ms</div></div></div>
      
      <!-- Throughput Metrics -->
      <div class="panel stat-panel col-2"><div class="panel-header"><div class="panel-title">Peak RPS</div></div><div class="panel-content"><div class="stat-value stat-green">{{ metrics.peakRPS }}</div><div class="stat-unit">req/sec</div></div></div>
      <div class="panel stat-panel col-2"><div class="panel-header"><div class="panel-title">Total Requests</div></div><div class="panel-content"><div class="stat-value stat-purple">{{ metrics.totalRequests }}</div></div></div>
      
      <!-- Quality Metrics -->
      <div class="panel stat-panel col-2"><div class="panel-header"><div class="panel-title">Error Rate</div></div><div class="panel-content"><div class="stat-value stat-red">{{ metrics.errorRate }}%</div></div></div>

      <!-- Charts -->
      <div class="panel col-12"><div class="panel-header"><div class="panel-title">Performance Timeline (Avg, P95, P99 Response)</div></div><div class="panel-content"><div class="chart-container medium-chart"><canvas ref="perfTimelineCanvas"></canvas></div></div></div>
      <div class="panel col-7"><div class="panel-header"><div class="panel-title">RPS Timeline</div></div><div class="panel-content"><div class="chart-container medium-chart"><canvas ref="rpsTimelineCanvas"></canvas></div></div></div>
      <div class="panel col-5"><div class="panel-header"><div class="panel-title">Response Time Distribution</div></div><div class="panel-content"><div class="chart-container medium-chart"><canvas ref="responseDistCanvas"></canvas></div></div></div>
      <div class="panel col-6"><div class="panel-header"><div class="panel-title">Top 10 Slowest Endpoints</div></div><div class="panel-content"><div class="chart-container medium-chart"><canvas ref="slowestEndpointsCanvas"></canvas></div></div></div>
      <div class="panel col-6"><div class="panel-header"><div class="panel-title">Top 10 Error Endpoints</div></div><div class="panel-content"><div class="chart-container medium-chart"><canvas ref="errorEndpointsCanvas"></canvas></div></div></div>
      <div class="panel col-4"><div class="panel-header"><div class="panel-title">Status Codes</div></div><div class="panel-content"><div class="chart-container small-chart"><canvas ref="statusCanvas"></canvas></div></div></div>
      <div class="panel col-4"><div class="panel-header"><div class="panel-title">HTTP Methods</div></div><div class="panel-content"><div class="chart-container small-chart"><canvas ref="methodCanvas"></canvas></div></div></div>
      <div class="panel col-4"><div class="panel-header"><div class="panel-title">Traffic by Hour</div></div><div class="panel-content"><div class="chart-container small-chart"><canvas ref="trafficHourCanvas"></canvas></div></div></div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue';
import Chart from 'chart.js/auto';
import 'chartjs-adapter-date-fns';
import { useApi } from '../service/api';

// State
const historyData = ref([]);
const filters = ref({ timeRange: '24h', endpoint: 'all' });
const api = useApi();

// Canvas Refs
const perfTimelineCanvas = ref(null);
const rpsTimelineCanvas = ref(null);
const responseDistCanvas = ref(null);
const statusCanvas = ref(null);
const methodCanvas = ref(null);
const trafficHourCanvas = ref(null);
const slowestEndpointsCanvas = ref(null);
const errorEndpointsCanvas = ref(null);

// Chart Instances
let perfTimelineChart, rpsTimelineChart, responseDistChart, statusChart, methodChart, trafficHourChart, slowestEndpointsChart, errorEndpointsChart;

// Computed Properties
const uniqueEndpoints = computed(() => [...new Set(historyData.value.map(d => d.endpoint))]);

const metrics = computed(() => {
  const data = historyData.value;
  if (!data.length) return { totalRequests: 0, avgResponse: 0, p95Response: 0, p99Response: 0, errorRate: 0, peakRPS: 0 };
  
  const total = data.length;
  const responseTimes = data.map(d => d.response_time).sort((a, b) => a - b);
  
  const rpsBySec = {};
  data.forEach(d => {
    const sec = Math.floor(new Date(d.timestamp || Date.now()).getTime() / 1000);
    rpsBySec[sec] = (rpsBySec[sec] || 0) + 1;
  });

  return {
    totalRequests: total,
    avgResponse: (responseTimes.reduce((s, t) => s + t, 0) / total).toFixed(1),
    p95Response: (responseTimes[Math.floor(total * 0.95)] || 0).toFixed(1),
    p99Response: (responseTimes[Math.floor(total * 0.99)] || 0).toFixed(1),
    errorRate: ((data.filter(d => d.status_code >= 400).length / total) * 100).toFixed(1),
    peakRPS: Math.max(0, ...Object.values(rpsBySec)),
  };
});

// Data Loading
async function loadDetailedData() {
  try {
    const response = await api.getHistory({ timeRange: filters.value.timeRange, endpoint: filters.value.endpoint });
    historyData.value = response.data || [];
    await nextTick();
    updateAllCharts();
  } catch (error) {
    console.error('Failed to load detailed data:', error);
  }
}

// Charting
function initCharts() {
  Chart.defaults.color = '#d9d9d9';
  Chart.defaults.font.family = "'Roboto', sans-serif";
  const commonOptions = { responsive: true, maintainAspectRatio: false };
  const timeScale = { x: { type: 'time', time: { unit: 'hour' } }, y: { beginAtZero: true } };

  perfTimelineChart = new Chart(perfTimelineCanvas.value, { type: 'line', options: { ...commonOptions, scales: timeScale } });
  rpsTimelineChart = new Chart(rpsTimelineCanvas.value, { type: 'line', options: { ...commonOptions, scales: timeScale } });
  responseDistChart = new Chart(responseDistCanvas.value, { type: 'bar', options: commonOptions });
  statusChart = new Chart(statusCanvas.value, { type: 'doughnut', options: commonOptions });
  methodChart = new Chart(methodCanvas.value, { type: 'pie', options: commonOptions });
  trafficHourChart = new Chart(trafficHourCanvas.value, { type: 'bar', options: commonOptions });
  slowestEndpointsChart = new Chart(slowestEndpointsCanvas.value, { type: 'bar', options: { ...commonOptions, indexAxis: 'y' } });
  errorEndpointsChart = new Chart(errorEndpointsCanvas.value, { type: 'bar', options: { ...commonOptions, indexAxis: 'y' } });
}

function updateAllCharts() {
    const data = historyData.value;
    
    // Generate complete time range based on filter
    const now = new Date();
    const timeRange = filters.value.timeRange;
    let hours = 24; // default
    switch(timeRange) {
        case '1h': hours = 1; break;
        case '6h': hours = 6; break;
        case '24h': hours = 24; break;
        case '168h': hours = 168; break; // 7 days
    }
    
    // Create complete hourly buckets with zero values
    const hourlyStats = {};
    const labels = [];
    
    for (let i = hours - 1; i >= 0; i--) {
        const hour = new Date(now.getTime() - i * 60 * 60 * 1000);
        const hourKey = hour.toISOString().slice(0, 13) + ':00:00.000Z';
        hourlyStats[hourKey] = { times: [], count: 0 };
        labels.push(hourKey);
    }
    
    // Fill with actual data
    data.forEach(d => {
        const hour = new Date(d.timestamp || Date.now()).toISOString().slice(0, 13) + ':00:00.000Z';
        if (hourlyStats[hour]) {
            hourlyStats[hour].times.push(d.response_time);
            hourlyStats[hour].count++;
        }
    });

    // Perf Timeline - handle zero values
    perfTimelineChart.data = { labels, datasets: [
        { 
            label: 'Avg', 
            data: labels.map(l => {
                const stats = hourlyStats[l];
                return stats.times.length > 0 ? stats.times.reduce((s, t) => s + t, 0) / stats.count : 0;
            }), 
            borderColor: '#52c5f7', 
            tension: 0.3, 
            pointRadius: 2,
            spanGaps: false
        },
        { 
            label: 'P95', 
            data: labels.map(l => {
                const stats = hourlyStats[l];
                if (stats.times.length === 0) return 0;
                const sorted = [...stats.times].sort((a,b) => a - b);
                return sorted[Math.floor(sorted.length * 0.95)] || 0;
            }), 
            borderColor: '#ffb74d', 
            tension: 0.3, 
            pointRadius: 2,
            spanGaps: false
        },
        { 
            label: 'P99', 
            data: labels.map(l => {
                const stats = hourlyStats[l];
                if (stats.times.length === 0) return 0;
                const sorted = [...stats.times].sort((a,b) => a - b);
                return sorted[Math.floor(sorted.length * 0.99)] || 0;
            }), 
            borderColor: '#f2495c', 
            tension: 0.3, 
            pointRadius: 2,
            spanGaps: false
        },
    ]};
    perfTimelineChart.update();

    // RPS Timeline - handle zero values
    rpsTimelineChart.data = { 
        labels, 
        datasets: [{ 
            label: 'RPS', 
            data: labels.map(l => hourlyStats[l].count / 3600), 
            borderColor: '#73bf69', 
            fill: true, 
            backgroundColor: 'rgba(115, 191, 105, 0.1)', 
            tension: 0.3,
            spanGaps: false
        }]
    };
    rpsTimelineChart.update();

    // Response Dist
    const buckets = [0, 50, 100, 250, 500, 1000, 2000, 5000];
    const bucketCounts = new Array(buckets.length).fill(0);
    if (data.length > 0) {
        data.forEach(d => {
            const bucketIndex = buckets.findIndex(b => d.response_time < b);
            bucketCounts[bucketIndex > -1 ? bucketIndex : buckets.length -1]++;
        });
    }
    responseDistChart.data = { labels: buckets.map((b, i) => i === 0 ? `<${b}ms` : `${buckets[i-1]}-${b}ms`), datasets: [{ label: 'Requests', data: bucketCounts, backgroundColor: '#73bf69' }]};
    responseDistChart.update();

    // Status & Method - handle empty data
    const statusCounts = data.length > 0 ? data.reduce((acc, d) => { acc[d.status_code] = (acc[d.status_code] || 0) + 1; return acc; }, {}) : {};
    const statusLabels = Object.keys(statusCounts).length > 0 ? Object.keys(statusCounts) : ['No Data'];
    const statusValues = Object.keys(statusCounts).length > 0 ? Object.values(statusCounts) : [1];
    statusChart.data = { labels: statusLabels, datasets: [{ data: statusValues, backgroundColor: ['#73bf69', '#fade2a', '#ffb74d', '#f2495c'] }]};
    statusChart.update();
    
    const methodCounts = data.length > 0 ? data.reduce((acc, d) => { acc[d.method] = (acc[d.method] || 0) + 1; return acc; }, {}) : {};
    const methodLabels = Object.keys(methodCounts).length > 0 ? Object.keys(methodCounts) : ['No Data'];
    const methodValues = Object.keys(methodCounts).length > 0 ? Object.values(methodCounts) : [1];
    methodChart.data = { labels: methodLabels, datasets: [{ data: methodValues, backgroundColor: ['#52c5f7', '#73bf69', '#ffb74d', '#f2495c', '#b877d9'] }]};
    methodChart.update();

    // Traffic by Hour - always show 24 hours with zeros
    const trafficByHour = new Array(24).fill(0);
    if (data.length > 0) {
        data.forEach(d => {
            const hour = new Date(d.timestamp || Date.now()).getHours();
            trafficByHour[hour]++;
        });
    }
    trafficHourChart.data = { labels: Array.from({length: 24}, (_, i) => `${i}:00`), datasets: [{ label: 'Requests', data: trafficByHour, backgroundColor: '#b877d9' }]};
    trafficHourChart.update();

    // Endpoint Charts - handle empty data
    let endpointStats = {};
    if (data.length > 0) {
        endpointStats = data.reduce((acc, d) => {
            const key = `${d.method} ${d.endpoint}`;
            if (!acc[key]) acc[key] = { count: 0, errors: 0, totalTime: 0 };
            acc[key].count++;
            acc[key].totalTime += d.response_time;
            if (d.status_code >= 400) acc[key].errors++;
            return acc;
        }, {});
    }

    const sortedByTime = Object.entries(endpointStats).sort(([,a],[,b]) => (b.totalTime/b.count) - (a.totalTime/a.count)).slice(0, 10);
    const slowestLabels = sortedByTime.length > 0 ? sortedByTime.map(([k]) => k) : ['No Data'];
    const slowestData = sortedByTime.length > 0 ? sortedByTime.map(([,v])=>v.totalTime/v.count) : [0];
    slowestEndpointsChart.data = { labels: slowestLabels, datasets: [{ label: 'Avg Response (ms)', data: slowestData, backgroundColor: '#ffb74d' }]};
    slowestEndpointsChart.update();

    const sortedByError = Object.entries(endpointStats).filter(([,v])=>v.errors > 0).sort(([,a],[,b]) => (b.errors/b.count) - (a.errors/a.count)).slice(0, 10);
    const errorLabels = sortedByError.length > 0 ? sortedByError.map(([k]) => k) : ['No Errors'];
    const errorData = sortedByError.length > 0 ? sortedByError.map(([,v])=>(v.errors/v.count)*100) : [0];
    errorEndpointsChart.data = { labels: errorLabels, datasets: [{ label: 'Error Rate (%)', data: errorData, backgroundColor: '#f2495c' }]};
    errorEndpointsChart.update();
}

// Lifecycle
onMounted(() => {
  initCharts();
  loadDetailedData();
  setInterval(loadDetailedData, 60000);
});
</script>

<style scoped>
.detailed-analytics-section {
  padding: 20px 0;
}
.panel.stat-panel.col-2 {
    grid-column: span 2;
}
/* Copy grid & panel styles... */
</style>
