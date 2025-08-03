<template>
  <div class="analytics-section">
    <!-- Page Header -->
    <PageHeader
      page-title="API Analytics Dashboard"
      page-description="Real-time monitoring and performance analysis of your API endpoints"
      page-icon="fas fa-tachometer-alt"
      :stats="headerStats"
      :refreshing="isRefreshing"
      :show-export="true"
      @refresh="handleRefresh"
      @export="handleExport"
    >
      <template #filters>
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
        <div class="analytics-filter-group">
          <label class="analytics-filter-label">Export</label>
          <button class="export-pdf-btn" @click="exportToPDF" :disabled="isExporting">
            <i class="fas fa-spinner fa-spin" v-if="isExporting"></i>
            <i class="fas fa-file-pdf" v-else></i>
            {{ isExporting ? 'Generating PDF...' : 'Export PDF Summary' }}
          </button>
        </div>
      </template>
    </PageHeader>

    <!-- Primary Health Overview -->
    <div class="dashboard-grid">
      <!-- Critical Metrics Row - Full Width Priority -->
      <div class="panel stat-panel critical-metric-large col-3">
        <div class="panel-header"><div class="panel-title">Error Rate</div></div>
        <div class="panel-content">
          <div class="stat-value-large stat-red">{{ qualityMetrics.errorRate }}%</div>
          <div class="stat-unit">4xx/5xx errors</div>
          <div class="stat-trend" :class="{ good: qualityMetrics.errorRate < 5, warning: qualityMetrics.errorRate >= 5 && qualityMetrics.errorRate < 10, critical: qualityMetrics.errorRate >= 10 }">
            {{ qualityMetrics.errorRate < 5 ? '✓ Healthy' : qualityMetrics.errorRate < 10 ? '⚠ Warning' : '✗ Critical' }}
          </div>
        </div>
      </div>

      <div class="panel stat-panel critical-metric-large col-3">
        <div class="panel-header"><div class="panel-title">Avg Response Time</div></div>
        <div class="panel-content">
          <div class="stat-value-large stat-blue">{{ responseMetrics.avgResponse }}ms</div>
          <div class="stat-unit">milliseconds</div>
          <div class="stat-trend" :class="{ good: responseMetrics.avgResponse < 200, warning: responseMetrics.avgResponse >= 200 && responseMetrics.avgResponse < 500, critical: responseMetrics.avgResponse >= 500 }">
            {{ responseMetrics.avgResponse < 200 ? '✓ Fast' : responseMetrics.avgResponse < 500 ? '⚠ Slow' : '✗ Very Slow' }}
          </div>
        </div>
      </div>

      <div class="panel stat-panel critical-metric-large col-3">
        <div class="panel-header"><div class="panel-title">Current RPS</div></div>
        <div class="panel-content">
          <div class="stat-value-large stat-green">{{ rpsMetrics.currentRPS }}</div>
          <div class="stat-unit">requests/sec</div>
          <div class="stat-trend">
            Peak: {{ rpsMetrics.peakRPS }} RPS
          </div>
        </div>
      </div>

      <div class="panel stat-panel critical-metric-large col-3">
        <div class="panel-header"><div class="panel-title">Success Rate</div></div>
        <div class="panel-content">
          <div class="stat-value-large stat-green">{{ qualityMetrics.successRate }}%</div>
          <div class="stat-unit">2xx responses</div>
          <div class="stat-trend" :class="{ good: qualityMetrics.successRate >= 95, warning: qualityMetrics.successRate >= 90 && qualityMetrics.successRate < 95, critical: qualityMetrics.successRate < 90 }">
            {{ qualityMetrics.successRate >= 95 ? '✓ Excellent' : qualityMetrics.successRate >= 90 ? '⚠ Good' : '✗ Poor' }}
          </div>
        </div>
      </div>

      <!-- Health Score & Quick Stats -->
      <div class="health-overview-section col-12">
        <div class="health-overview-container">
          <!-- Health Score -->
          <div class="health-score-compact">
            <div class="health-score-value-compact" :class="healthScoreClass">{{ apiHealthScore }}</div>
            <div class="health-score-label-compact">API Health Score</div>
          </div>
          
          <!-- Secondary Stats -->
          <div class="secondary-stats-grid">
            <div class="secondary-stat">
              <div class="secondary-stat-value">{{ responseMetrics.p95Response }}ms</div>
              <div class="secondary-stat-label">P95 Response</div>
            </div>
            <div class="secondary-stat">
              <div class="secondary-stat-value">{{ rpsMetrics.totalRequests.toLocaleString() }}</div>
              <div class="secondary-stat-label">Total Requests</div>
            </div>
            <div class="secondary-stat">
              <div class="secondary-stat-value">{{ rpsMetrics.peakRPS }}</div>
              <div class="secondary-stat-label">Peak RPS</div>
            </div>
            <div class="secondary-stat">
              <div class="secondary-stat-value">{{ activeEndpoints }}</div>
              <div class="secondary-stat-label">Active Endpoints</div>
            </div>
          </div>
          
          <!-- Health Indicators -->
          <div class="health-indicators-compact">
            <div class="health-indicator-compact" :class="{ good: qualityMetrics.successRate >= 95 }">
              <i class="fas fa-check-circle"></i>
              <span>Success: {{ qualityMetrics.successRate }}%</span>
            </div>
            <div class="health-indicator-compact" :class="{ good: responseMetrics.avgResponse <= 500 }">
              <i class="fas fa-clock"></i>
              <span>Avg RT: {{ responseMetrics.avgResponse }}ms</span>
            </div>
            <div class="health-indicator-compact" :class="{ good: qualityMetrics.errorRate <= 5 }">
              <i class="fas fa-exclamation-triangle"></i>
              <span>Errors: {{ qualityMetrics.errorRate }}%</span>
            </div>
          </div>
        </div>
      </div>



      <!-- Primary Performance Charts -->
      <div class="chart-section-header col-12">
        <h3>Performance & Request Trends</h3>
      </div>

      <!-- Main Timeline Charts -->
      <div class="panel chart-panel-primary col-12">
        <div class="panel-header">
          <div class="panel-title">Requests Per Second Timeline</div>
          <div class="panel-subtitle">Real-time request volume monitoring</div>
        </div>
        <div class="panel-content">
          <div class="chart-container time-series-chart">
            <canvas ref="rpsTimelineCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel chart-panel-primary col-12">
        <div class="panel-header">
          <div class="panel-title">Performance Timeline (Avg & P95 Response Time)</div>
          <div class="panel-subtitle">Response time average and 95th percentile over time</div>
        </div>
        <div class="panel-content">
          <div class="chart-container time-series-chart">
            <canvas ref="perfTimelineCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel chart-panel-primary col-12">
        <div class="panel-header">
          <div class="panel-title">Error Rate Timeline (4xx & 5xx Errors Over Time)</div>
          <div class="panel-subtitle">Error patterns and anomaly detection</div>
        </div>
        <div class="panel-content">
          <div class="chart-container time-series-chart">
            <canvas ref="errorTimelineCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Secondary Analysis Charts -->
      <div class="chart-section-header col-12">
        <h3>Detailed Analysis</h3>
      </div>

      <!-- Mid-level Charts -->
      <div class="panel chart-panel-secondary col-6">
        <div class="panel-header">
          <div class="panel-title">Average Response Time Trend</div>
        </div>
        <div class="panel-content">
          <div class="chart-container medium-chart">
            <canvas ref="avgResponseTimeCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel chart-panel-secondary col-6">
        <div class="panel-header">
          <div class="panel-title">Request Volume Trend</div>
        </div>
        <div class="panel-content">
          <div class="chart-container medium-chart">
            <canvas ref="totalRequestsCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Distribution Analysis -->
      <div class="chart-section-header col-12">
        <h3>Distribution Analysis</h3>
      </div>

      <!-- Distribution Charts Row -->
      <div class="panel chart-panel-tertiary col-4">
        <div class="panel-header">
          <div class="panel-title">Response Time Distribution</div>
        </div>
        <div class="panel-content">
          <div class="chart-container small-chart">
            <canvas ref="distributionCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel chart-panel-tertiary col-4">
        <div class="panel-header">
          <div class="panel-title">HTTP Status Distribution</div>
        </div>
        <div class="panel-content">
          <div class="chart-container small-chart">
            <canvas ref="statusDistributionCanvas"></canvas>
          </div>
        </div>
      </div>

      <div class="panel chart-panel-tertiary col-4">
        <div class="panel-header">
          <div class="panel-title">Method Distribution</div>
        </div>
        <div class="panel-content">
          <div class="chart-container small-chart">
            <canvas ref="methodDistributionCanvas"></canvas>
          </div>
        </div>
      </div>

      <!-- Endpoint Analysis -->
      <div class="chart-section-header col-12">
        <h3>Endpoint Performance Analysis</h3>
      </div>

      <div class="panel chart-panel-secondary col-12">
        <div class="panel-header">
          <div class="panel-title">Response Time by Endpoint (Top 10 Slowest)</div>
          <div class="panel-subtitle">Identify performance bottlenecks by endpoint</div>
        </div>
        <div class="panel-content">
          <div class="chart-container endpoint-chart">
            <canvas ref="endpointCanvas"></canvas>
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

const historyData = ref([]);
const isRefreshing = ref(false);
const filters = ref({
  timeRange: '24h',
  endpoint: 'all',
  method: 'all'
});

const isExporting = ref(false);

const rpsTimelineCanvas = ref(null);
const perfTimelineCanvas = ref(null);
const distributionCanvas = ref(null);
const statusDistributionCanvas = ref(null);
const methodDistributionCanvas = ref(null);
const endpointCanvas = ref(null);
const avgResponseTimeCanvas = ref(null);
const totalRequestsCanvas = ref(null);
const errorTimelineCanvas = ref(null);

let rpsTimelineChart, perfTimelineChart, distributionChart, statusDistributionChart, methodDistributionChart, endpointChart, avgResponseTimeChart, totalRequestsChart, errorTimelineChart;
const api = useApi();

// Computed properties
const uniqueEndpoints = computed(() => {
  const data = historyData.value;
  if (!data || !data.length) return [];
  return [...new Set(data.map(d => d.endpoint).filter(Boolean))];
});

const responseMetrics = computed(() => {
  const data = historyData.value;
  if (!data || !data.length) return { avgResponse: '0', p95Response: '0' };
  
  const responseTimes = data.filter(d => d.response_time && typeof d.response_time === 'number').map(d => d.response_time);
  if (!responseTimes.length) return { avgResponse: '0', p95Response: '0' };
  
  responseTimes.sort((a, b) => a - b);
  const avgResponse = (responseTimes.reduce((sum, t) => sum + t, 0) / responseTimes.length).toFixed(1);
  const p95Index = Math.floor(responseTimes.length * 0.95);
  const p95Response = responseTimes[p95Index]?.toFixed(1) || '0';
  
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

const activeEndpoints = computed(() => {
  const data = historyData.value;
  if (!data || !data.length) return 0;
  return new Set(data.map(d => `${d.method} ${d.endpoint}`)).size;
});

const qualityMetrics = computed(() => {
  const data = historyData.value;
  if (!data || !data.length) return { errorRate: '0', successRate: '100' };
  
  const totalRequests = data.length;
  const successRequests = data.filter(d => d.status_code >= 200 && d.status_code < 400).length;
  const errorRequests = data.filter(d => d.status_code >= 400).length;
  
  const errorRate = ((errorRequests / totalRequests) * 100).toFixed(1);
  const successRate = ((successRequests / totalRequests) * 100).toFixed(1);
  
  return { errorRate, successRate };
});

const methodMetrics = computed(() => {
  const data = historyData.value;
  if (!data || !data.length) return {};
  
  const methodCounts = {};
  data.forEach(d => {
    if (d.method) {
      methodCounts[d.method] = (methodCounts[d.method] || 0) + 1;
    }
  });
  
  return methodCounts;
});

const apiHealthScore = computed(() => {
  const { successRate, errorRate } = qualityMetrics.value;
  const { avgResponse } = responseMetrics.value;
  
  let score = 100;
  score -= (100 - parseFloat(successRate || '100')) * 0.5;
  score -= parseFloat(errorRate || '0') * 0.3;
  score -= Math.max(0, (parseFloat(avgResponse || '0') - 200) / 10);
  
  return Math.max(0, Math.min(100, score)).toFixed(0);
});

const healthScoreClass = computed(() => {
  const score = parseFloat(apiHealthScore.value);
  if (score >= 90) return 'excellent';
  if (score >= 70) return 'good';
  if (score >= 50) return 'warning';
  return 'critical';
});

// Header stats for PageHeader component
const headerStats = computed(() => {
  const healthScore = apiHealthScore.value || '0';
  const errorRate = qualityMetrics.value.errorRate || '0';
  const avgResponse = responseMetrics.value.avgResponse || '0';
  const totalRequests = rpsMetrics.value.totalRequests || 0;
  
  return [
    {
      label: 'Health Score',
      value: healthScore,
      type: parseFloat(healthScore) >= 90 ? 'success' : parseFloat(healthScore) >= 70 ? 'warning' : 'critical'
    },
    {
      label: 'Error Rate',
      value: `${errorRate}%`,
      type: parseFloat(errorRate) < 5 ? 'success' : parseFloat(errorRate) < 10 ? 'warning' : 'critical'
    },
    {
      label: 'Avg Response',
      value: `${avgResponse}ms`,
      type: parseFloat(avgResponse) < 200 ? 'success' : parseFloat(avgResponse) < 500 ? 'warning' : 'critical'
    },
    {
      label: 'Total Requests',
      value: totalRequests.toLocaleString(),
      type: 'info'
    }
  ];
});

// Handler functions
const handleRefresh = async () => {
  isRefreshing.value = true;
  await loadAnalyticsData();
  isRefreshing.value = false;
};

const handleExport = () => {
  // Implement export functionality
  const data = {
    timestamp: new Date().toISOString(),
    filters: filters.value,
    metrics: {
      response: responseMetrics.value,
      rps: rpsMetrics.value,
      quality: qualityMetrics.value,
      activeEndpoints: activeEndpoints.value,
      health: apiHealthScore.value
    },
    data: historyData.value
  };
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `api-analytics-${new Date().toISOString().split('T')[0]}.json`;
  a.click();
  URL.revokeObjectURL(url);
};

// Export to PDF functionality
const exportToPDF = async () => {
  if (isExporting.value) return;
  
  isExporting.value = true;
  
  try {
    // Load jsPDF dynamically
    const { jsPDF } = await import('jspdf');
    
    // Use the currently filtered data from the dashboard
    const summaryData = generateSummaryData(historyData.value);
    
    // Create new PDF document
    const doc = new jsPDF();
    
    // Set document properties
    doc.setProperties({
      title: 'API Analytics Summary Report',
      subject: 'Performance and Error Analysis',
      author: 'HealthTrack API Monitor',
      creator: 'API Analytics Dashboard'
    });
    
    // Generate PDF content
    await generatePDFContent(doc, summaryData);
    
    // Generate filename based on active filter
    const timeRangeText = filters.value.timeRange.replace('h', '-hours');
    const filename = `api-summary-${timeRangeText}-${new Date().toISOString().split('T')[0]}.pdf`;
    doc.save(filename);
    
    console.log('PDF exported successfully:', filename);
    
  } catch (error) {
    console.error('Failed to export PDF:', error);
    alert('Failed to export PDF. Please try again.');
  } finally {
    isExporting.value = false;
  }
};

// Get filtered data for export
const getFilteredDataForExport = async () => {
  try {
    const params = {
      startDate: exportFilters.value.startDate,
      endDate: exportFilters.value.endDate
    };
    
    const response = await api.getHistory(params);
    return response.data || [];
  } catch (error) {
    console.error('Failed to load export data:', error);
    return [];
  }
};

// Generate summary data for PDF
const generateSummaryData = (data) => {
  if (!data || !data.length) {
    return {
      totalRequests: 0,
      errorRate: '0',
      successRate: '100',
      avgResponseTime: '0',
      p95ResponseTime: '0',
      slowestEndpoints: [],
      errorsByStatus: {},
      requestsByMethod: {},
      requestsByDay: {},
      healthScore: '0'
    };
  }
  
  // Calculate metrics
  const totalRequests = data.length;
  const errorRequests = data.filter(d => d.status_code >= 400).length;
  const successRequests = data.filter(d => d.status_code >= 200 && d.status_code < 400).length;
  
  const errorRate = ((errorRequests / totalRequests) * 100).toFixed(1);
  const successRate = ((successRequests / totalRequests) * 100).toFixed(1);
  
  // Response time metrics
  const responseTimes = data.map(d => d.response_time).sort((a, b) => a - b);
  const avgResponseTime = (responseTimes.reduce((sum, t) => sum + t, 0) / responseTimes.length).toFixed(1);
  const p95Index = Math.floor(responseTimes.length * 0.95);
  const p95ResponseTime = responseTimes[p95Index]?.toFixed(1) || 0;
  
  // Slowest endpoints
  const endpointStats = {};
  data.forEach(d => {
    const key = `${d.method} ${d.endpoint}`;
    if (!endpointStats[key]) {
      endpointStats[key] = { count: 0, totalResponseTime: 0 };
    }
    endpointStats[key].count++;
    endpointStats[key].totalResponseTime += d.response_time;
  });
  
  const slowestEndpoints = Object.entries(endpointStats)
    .map(([endpoint, stats]) => ({
      endpoint,
      avgResponse: (stats.totalResponseTime / stats.count).toFixed(1),
      count: stats.count
    }))
    .sort((a, b) => parseFloat(b.avgResponse) - parseFloat(a.avgResponse))
    .slice(0, 5);
  
  // Errors by status
  const errorsByStatus = {};
  data.forEach(d => {
    if (d.status_code >= 400) {
      const statusGroup = Math.floor(d.status_code / 100) + 'xx';
      errorsByStatus[statusGroup] = (errorsByStatus[statusGroup] || 0) + 1;
    }
  });
  
  // Requests by method
  const requestsByMethod = {};
  data.forEach(d => {
    requestsByMethod[d.method] = (requestsByMethod[d.method] || 0) + 1;
  });
  
  // Requests by day
  const requestsByDay = {};
  data.forEach(d => {
    const day = new Date(d.timestamp).toISOString().split('T')[0];
    requestsByDay[day] = (requestsByDay[day] || 0) + 1;
  });
  
  // Health score calculation
  let healthScore = 100;
  healthScore -= (100 - parseFloat(successRate)) * 0.5;
  healthScore -= parseFloat(errorRate) * 0.3;
  healthScore -= Math.max(0, (parseFloat(avgResponseTime) - 200) / 10);
  healthScore = Math.max(0, Math.min(100, healthScore)).toFixed(0);
  
  return {
    totalRequests,
    errorRate,
    successRate,
    avgResponseTime,
    p95ResponseTime,
    slowestEndpoints,
    errorsByStatus,
    requestsByMethod,
    requestsByDay,
    healthScore
  };
};

// Generate PDF content
const generatePDFContent = async (doc, summaryData) => {
  const pageWidth = doc.internal.pageSize.width;
  const pageHeight = doc.internal.pageSize.height;
  let yPosition = 20;
  
  // Header
  doc.setFontSize(20);
  doc.setFont(undefined, 'bold');
  doc.text('API Analytics Summary Report', pageWidth / 2, yPosition, { align: 'center' });
  
  yPosition += 10;
  doc.setFontSize(12);
  doc.setFont(undefined, 'normal');
  const timeRangeLabel = `Period: Last ${filters.value.timeRange.replace('h', ' Hours')}`;
  doc.text(timeRangeLabel, pageWidth / 2, yPosition, { align: 'center' });
  doc.text(`Generated: ${new Date().toLocaleString()}`, pageWidth / 2, yPosition + 5, { align: 'center' });
  
  yPosition += 25;
  
  // Executive Summary
  doc.setFontSize(16);
  doc.setFont(undefined, 'bold');
  doc.text('Executive Summary', 20, yPosition);
  yPosition += 10;
  
  doc.setFontSize(11);
  doc.setFont(undefined, 'normal');
  
  const summaryLines = [
    `Total Requests: ${summaryData.totalRequests.toLocaleString()}`,
    `Success Rate: ${summaryData.successRate}%`,
    `Error Rate: ${summaryData.errorRate}%`,
    `Average Response Time: ${summaryData.avgResponseTime}ms`,
    `P95 Response Time: ${summaryData.p95ResponseTime}ms`,
    `API Health Score: ${summaryData.healthScore}/100`
  ];
  
  summaryLines.forEach(line => {
    doc.text(line, 25, yPosition);
    yPosition += 7;
  });
  
  yPosition += 10;
  
  // Performance Analysis
  doc.setFontSize(14);
  doc.setFont(undefined, 'bold');
  doc.text('Performance Analysis', 20, yPosition);
  yPosition += 10;
  
  doc.setFontSize(11);
  doc.setFont(undefined, 'normal');
  doc.text('Top 5 Slowest Endpoints:', 25, yPosition);
  yPosition += 7;
  
  summaryData.slowestEndpoints.forEach((endpoint, index) => {
    const text = `${index + 1}. ${endpoint.endpoint} - ${endpoint.avgResponse}ms (${endpoint.count} requests)`;
    doc.text(text, 30, yPosition);
    yPosition += 6;
  });
  
  yPosition += 10;
  
  // Error Analysis
  doc.setFontSize(14);
  doc.setFont(undefined, 'bold');
  doc.text('Error Analysis', 20, yPosition);
  yPosition += 10;
  
  doc.setFontSize(11);
  doc.setFont(undefined, 'normal');
  
  if (Object.keys(summaryData.errorsByStatus).length > 0) {
    doc.text('Errors by Status Code:', 25, yPosition);
    yPosition += 7;
    
    Object.entries(summaryData.errorsByStatus).forEach(([status, count]) => {
      doc.text(`${status}: ${count} errors`, 30, yPosition);
      yPosition += 6;
    });
  } else {
    doc.text('No errors detected in this period.', 25, yPosition);
    yPosition += 7;
  }
  
  yPosition += 10;
  
  // Request Distribution
  doc.setFontSize(14);
  doc.setFont(undefined, 'bold');
  doc.text('Request Distribution', 20, yPosition);
  yPosition += 10;
  
  doc.setFontSize(11);
  doc.setFont(undefined, 'normal');
  doc.text('Requests by HTTP Method:', 25, yPosition);
  yPosition += 7;
  
  Object.entries(summaryData.requestsByMethod).forEach(([method, count]) => {
    const percentage = ((count / summaryData.totalRequests) * 100).toFixed(1);
    doc.text(`${method}: ${count} (${percentage}%)`, 30, yPosition);
    yPosition += 6;
  });
  
  yPosition += 10;
  
  // Daily breakdown (if we have space)
  if (yPosition < pageHeight - 60) {
    doc.setFontSize(14);
    doc.setFont(undefined, 'bold');
    doc.text('Daily Request Volume', 20, yPosition);
    yPosition += 10;
    
    doc.setFontSize(11);
    doc.setFont(undefined, 'normal');
    
    const sortedDays = Object.entries(summaryData.requestsByDay)
      .sort(([a], [b]) => a.localeCompare(b));
      
    sortedDays.forEach(([day, count]) => {
      doc.text(`${day}: ${count} requests`, 25, yPosition);
      yPosition += 6;
      
      if (yPosition > pageHeight - 30) {
        doc.addPage();
        yPosition = 20;
      }
    });
  }
  
  // Footer
  doc.setFontSize(8);
  doc.setFont(undefined, 'italic');
  doc.text('Generated by HealthTrack API Analytics Dashboard', pageWidth / 2, pageHeight - 10, { align: 'center' });
};

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
    
    console.log('Loaded analytics data:', historyData.value.length, 'records');
    
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

  perfTimelineChart = new Chart(perfTimelineCanvas.value, {
    type: 'line',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { type: 'time', time: { unit: 'minute' } },
        y: { beginAtZero: true, title: { display: true, text: 'Response Time (ms)' } }
      },
      plugins: {
        legend: {
          display: true,
          position: 'top'
        }
      }
    }
  });

  errorTimelineChart = new Chart(errorTimelineCanvas.value, {
    type: 'line',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { type: 'time', time: { unit: 'minute' } },
        y: { 
          beginAtZero: true, 
          max: 100,
          title: { display: true, text: 'Error Rate (%)' },
          ticks: {
            callback: function(value) {
              return value + '%';
            }
          }
        }
      },
      plugins: {
        legend: {
          display: true,
          position: 'top'
        }
      }
    }
  });

  avgResponseTimeChart = new Chart(avgResponseTimeCanvas.value, {
    type: 'line',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { type: 'time', time: { unit: 'minute' } },
        y: { beginAtZero: true, title: { display: true, text: 'Average Response Time (ms)' } }
      }
    }
  });

  totalRequestsChart = new Chart(totalRequestsCanvas.value, {
    type: 'line',
    data: { labels: [], datasets: [] },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        x: { type: 'time', time: { unit: 'minute' } },
        y: { beginAtZero: true, title: { display: true, text: 'Total Requests' } }
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

  methodDistributionChart = new Chart(methodDistributionCanvas.value, {
    type: 'doughnut',
    data: { labels: [], datasets: [] },
    options: { 
      responsive: true, 
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'bottom',
          labels: {
            padding: 10,
            font: { size: 11 }
          }
        }
      }
    }
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
  if (!data || !Array.isArray(data)) {
    console.warn('No valid data available for charts');
    return;
  }
  
  // Common time range variables
  const timeRangeMap = { '1h': 60, '6h': 360, '24h': 1440, '168h': 10080 }; // minutes
  const minutes = timeRangeMap[filters.value.timeRange] || 1440;
  const now = new Date();
  
  // RPS Timeline - with zero values for missing periods
  const minutelyRPS = {};
  data.forEach(d => {
    if (d.timestamp) {
      const minute = new Date(d.timestamp).toISOString().slice(0, 16) + ':00.000Z';
      minutelyRPS[minute] = (minutelyRPS[minute] || 0) + 1;
    }
  });

  // Generate complete time range with zero values for missing periods
  const rpsLabels = [];
  const rpsValues = [];

  for (let i = minutes; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60000);
    const timeKey = time.toISOString().slice(0, 16) + ':00.000Z';
    rpsLabels.push(timeKey);
    rpsValues.push(((minutelyRPS[timeKey] || 0) / 60).toFixed(2));
  }

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

  // Performance Timeline - with Avg, P95 response times
  const perfTimelineData = {};
  data.forEach(d => {
    if (d.timestamp && d.response_time && typeof d.response_time === 'number') {
      const minute = new Date(d.timestamp).toISOString().slice(0, 16) + ':00.000Z';
      if (!perfTimelineData[minute]) {
        perfTimelineData[minute] = [];
      }
      perfTimelineData[minute].push(d.response_time);
    }
  });

  // Generate complete time range with calculated percentiles
  const perfLabels = [];
  const avgValues = [];
  const p95Values = [];

  for (let i = minutes; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60000);
    const timeKey = time.toISOString().slice(0, 16) + ':00.000Z';
    perfLabels.push(timeKey);
    
    if (perfTimelineData[timeKey] && perfTimelineData[timeKey].length > 0) {
      const times = perfTimelineData[timeKey].sort((a, b) => a - b);
      const avg = times.reduce((sum, t) => sum + t, 0) / times.length;
      const p95Index = Math.floor(times.length * 0.95);
      
      avgValues.push(avg.toFixed(1));
      p95Values.push(times[p95Index]?.toFixed(1) || 0);
    } else {
      avgValues.push(0);
      p95Values.push(0);
    }
  }

  perfTimelineChart.data.labels = perfLabels;
  perfTimelineChart.data.datasets = [
    {
      label: 'Average',
      data: avgValues,
      borderColor: '#52c5f7',
      backgroundColor: 'rgba(82, 197, 247, 0.1)',
      fill: false,
      tension: 0.4,
      pointRadius: 2
    },
    {
      label: 'P95',
      data: p95Values,
      borderColor: '#ffb74d',
      backgroundColor: 'rgba(255, 183, 77, 0.1)',
      fill: false,
      tension: 0.4,
      pointRadius: 2
    }
  ];
  perfTimelineChart.update();

  // Error Timeline - with 4xx and 5xx error rates over time
  const errorTimelineData = {};
  data.forEach(d => {
    if (d.timestamp && d.status_code) {
      const minute = new Date(d.timestamp).toISOString().slice(0, 16) + ':00.000Z';
      if (!errorTimelineData[minute]) {
        errorTimelineData[minute] = { total: 0, errors4xx: 0, errors5xx: 0 };
      }
      errorTimelineData[minute].total++;
      
      if (d.status_code >= 400 && d.status_code < 500) {
        errorTimelineData[minute].errors4xx++;
      } else if (d.status_code >= 500) {
        errorTimelineData[minute].errors5xx++;
      }
    }
  });

  // Generate complete time range with calculated error rates
  const errorLabels = [];
  const error4xxValues = [];
  const error5xxValues = [];
  const totalErrorValues = [];

  for (let i = minutes; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60000);
    const timeKey = time.toISOString().slice(0, 16) + ':00.000Z';
    errorLabels.push(timeKey);
    
    if (errorTimelineData[timeKey] && errorTimelineData[timeKey].total > 0) {
      const total = errorTimelineData[timeKey].total;
      const errors4xx = errorTimelineData[timeKey].errors4xx;
      const errors5xx = errorTimelineData[timeKey].errors5xx;
      const totalErrors = errors4xx + errors5xx;
      
      error4xxValues.push(((errors4xx / total) * 100).toFixed(2));
      error5xxValues.push(((errors5xx / total) * 100).toFixed(2));
      totalErrorValues.push(((totalErrors / total) * 100).toFixed(2));
    } else {
      error4xxValues.push(0);
      error5xxValues.push(0);
      totalErrorValues.push(0);
    }
  }

  errorTimelineChart.data.labels = errorLabels;
  errorTimelineChart.data.datasets = [
    {
      label: 'Total Errors',
      data: totalErrorValues,
      borderColor: '#f2495c',
      backgroundColor: 'rgba(242, 73, 92, 0.1)',
      fill: true,
      tension: 0.4,
      pointRadius: 2
    },
    {
      label: '4xx Errors',
      data: error4xxValues,
      borderColor: '#ffb74d',
      backgroundColor: 'rgba(255, 183, 77, 0.1)',
      fill: false,
      tension: 0.4,
      pointRadius: 2
    },
    {
      label: '5xx Errors',
      data: error5xxValues,
      borderColor: '#b877d9',
      backgroundColor: 'rgba(184, 119, 217, 0.1)',
      fill: false,
      tension: 0.4,
      pointRadius: 2
    }
  ];
  errorTimelineChart.update();

  // Average Response Time Over Time - with zero values for missing periods
  const responseTimeData = {};
  data.forEach(d => {
    if (d.timestamp && d.response_time && typeof d.response_time === 'number') {
      const minute = new Date(d.timestamp).toISOString().slice(0, 16) + ':00.000Z';
      if (!responseTimeData[minute]) {
        responseTimeData[minute] = { totalResponseTime: 0, count: 0 };
      }
      responseTimeData[minute].totalResponseTime += d.response_time;
      responseTimeData[minute].count++;
    }
  });

  // Generate complete time range with zero values for missing periods
  const avgResponseLabels = [];
  const avgResponseValues = [];

  for (let i = minutes; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60000);
    const timeKey = time.toISOString().slice(0, 16) + ':00.000Z';
    avgResponseLabels.push(timeKey);
    
    if (responseTimeData[timeKey] && responseTimeData[timeKey].count > 0) {
      avgResponseValues.push((responseTimeData[timeKey].totalResponseTime / responseTimeData[timeKey].count).toFixed(1));
    } else {
      avgResponseValues.push(0);
    }
  }

  avgResponseTimeChart.data.labels = avgResponseLabels;
  avgResponseTimeChart.data.datasets = [{
    label: 'Avg Response Time (ms)',
    data: avgResponseValues,
    borderColor: '#ffb74d',
    backgroundColor: 'rgba(255, 183, 77, 0.1)',
    fill: true,
    tension: 0.4
  }];
  avgResponseTimeChart.update();

  // Total Requests Over Time - with zero values for missing periods
  const totalRequestsData = {};
  data.forEach(d => {
    if (d.timestamp) {
      const minute = new Date(d.timestamp).toISOString().slice(0, 16) + ':00.000Z';
      totalRequestsData[minute] = (totalRequestsData[minute] || 0) + 1;
    }
  });

  // Generate complete time range with zero values for missing periods
  const totalRequestsLabels = [];
  const totalRequestsValues = [];

  for (let i = minutes; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 60000);
    const timeKey = time.toISOString().slice(0, 16) + ':00.000Z';
    totalRequestsLabels.push(timeKey);
    totalRequestsValues.push(totalRequestsData[timeKey] || 0);
  }

  totalRequestsChart.data.labels = totalRequestsLabels;
  totalRequestsChart.data.datasets = [{
    label: 'Total Requests',
    data: totalRequestsValues,
    borderColor: '#73bf69',
    backgroundColor: 'rgba(115, 191, 105, 0.1)',
    fill: true,
    tension: 0.4
  }];
  totalRequestsChart.update();

  // Response Time Distribution
  const buckets = [0, 100, 250, 500, 1000, 2000, 5000];
  const bucketCounts = new Array(buckets.length - 1).fill(0);
  const bucketLabels = buckets.slice(0, -1).map((b, i) => `${b}-${buckets[i + 1]}ms`);

  data.forEach(d => {
    if (d.response_time && typeof d.response_time === 'number') {
      for (let i = 0; i < buckets.length - 1; i++) {
        if (d.response_time >= buckets[i] && d.response_time < buckets[i + 1]) {
          bucketCounts[i]++;
          break;
        }
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
    if (d.status_code) {
      const statusGroup = Math.floor(d.status_code / 100) + 'xx';
      statusCounts[statusGroup] = (statusCounts[statusGroup] || 0) + 1;
    }
  });

  if (Object.keys(statusCounts).length > 0) {
    statusDistributionChart.data.labels = Object.keys(statusCounts);
    statusDistributionChart.data.datasets = [{
      data: Object.values(statusCounts),
      backgroundColor: ['#73bf69', '#fade2a', '#ffb74d', '#f2495c', '#b877d9']
    }];
  } else {
    statusDistributionChart.data.labels = ['No Data'];
    statusDistributionChart.data.datasets = [{
      data: [1],
      backgroundColor: ['#9ca3af']
    }];
  }
  statusDistributionChart.update();

  // Method Distribution
  const methodCounts = methodMetrics.value;
  const methodColors = {
    'GET': '#52c5f7',
    'POST': '#73bf69',
    'PUT': '#ffb74d',
    'DELETE': '#f2495c',
    'PATCH': '#b877d9',
    'HEAD': '#fade2a',
    'OPTIONS': '#9ca3af'
  };

  if (Object.keys(methodCounts).length > 0) {
    methodDistributionChart.data.labels = Object.keys(methodCounts);
    methodDistributionChart.data.datasets = [{
      data: Object.values(methodCounts),
      backgroundColor: Object.keys(methodCounts).map(method => methodColors[method] || '#9ca3af')
    }];
  } else {
    methodDistributionChart.data.labels = ['No Data'];
    methodDistributionChart.data.datasets = [{
      data: [1],
      backgroundColor: ['#9ca3af']
    }];
  }
  methodDistributionChart.update();

  // Endpoint Chart
  const endpointStats = {};
  data.forEach(d => {
    if (d.method && d.endpoint && d.response_time) {
      const key = `${d.method} ${d.endpoint}`;
      if (!endpointStats[key]) {
        endpointStats[key] = { count: 0, totalResponseTime: 0 };
      }
      endpointStats[key].count++;
      endpointStats[key].totalResponseTime += d.response_time;
    }
  });

  const topEndpoints = Object.entries(endpointStats)
    .map(([endpoint, stats]) => ({
      endpoint: endpoint.length > 40 ? endpoint.substring(0, 37) + '...' : endpoint,
      avgResponse: (stats.totalResponseTime / stats.count).toFixed(1)
    }))
    .sort((a, b) => parseFloat(b.avgResponse) - parseFloat(a.avgResponse))
    .slice(0, 10);

  if (topEndpoints.length > 0) {
    endpointChart.data.labels = topEndpoints.map(e => e.endpoint);
    endpointChart.data.datasets = [{
      label: 'Avg Response Time (ms)',
      data: topEndpoints.map(e => parseFloat(e.avgResponse)),
      backgroundColor: 'rgba(255, 183, 77, 0.8)',
      borderColor: '#ffb74d',
      borderWidth: 1
    }];
  } else {
    endpointChart.data.labels = ['No Data Available'];
    endpointChart.data.datasets = [{
      label: 'Avg Response Time (ms)',
      data: [0],
      backgroundColor: 'rgba(156, 163, 175, 0.8)',
      borderColor: '#9ca3af',
      borderWidth: 1
    }];
  }
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

.analytics-filter-group {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.analytics-filter-label {
  font-size: 12px;
  color: var(--grafana-text-muted);
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.analytics-select {
  padding: 8px 12px;
  background-color: var(--grafana-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
  color: var(--grafana-text);
  font-size: 14px;
  min-width: 120px;
  transition: border-color 0.2s ease;
}

.analytics-select:focus {
  outline: none;
  border-color: var(--grafana-blue);
}

/* Dashboard Grid Improvements */
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

/* Section Headers */
.metrics-section-header,
.chart-section-header {
  grid-column: span 12;
  margin: 30px 0 15px 0;
  padding: 0;
  background: none;
  border: none;
}

.metrics-section-header h3,
.chart-section-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--grafana-text);
  margin: 0;
  padding: 12px 0;
  border-bottom: 2px solid var(--grafana-border);
  position: relative;
}

.metrics-section-header:first-of-type {
  margin-top: 0;
}

/* Primary Health Panel - Enhanced */
.primary-health-panel {
  background: linear-gradient(135deg, var(--grafana-panel-bg) 0%, rgba(82, 197, 247, 0.05) 100%);
  border: 2px solid var(--grafana-blue);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  position: relative;
}

.primary-health-panel::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--grafana-blue), var(--grafana-green));
}

/* Critical Metrics - Large Size for Full Screen */
.critical-metric-large {
  border-left: 6px solid var(--grafana-blue);
  background: linear-gradient(135deg, var(--grafana-panel-bg) 0%, rgba(255, 255, 255, 0.03) 100%);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  position: relative;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  min-height: 160px;
}

.critical-metric-large:hover {
  transform: translateY(-3px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.2);
}

.critical-metric-large .panel-header {
  background: linear-gradient(90deg, var(--grafana-bg-light), transparent);
  border-bottom: 1px solid var(--grafana-border);
  padding: 12px 16px;
}

.critical-metric-large .panel-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--grafana-text);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.critical-metric-large .panel-content {
  padding: 24px 20px;
}

.stat-value-large {
  font-size: 42px;
  font-weight: 900;
  margin-bottom: 6px;
  line-height: 1;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.critical-metric-large .stat-unit {
  font-size: 12px;
  color: var(--grafana-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 12px;
  font-weight: 500;
}

.critical-metric-large .stat-trend {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  padding: 6px 12px;
  border-radius: 15px;
  background-color: var(--grafana-bg-light);
  display: inline-block;
  min-width: 100px;
  text-align: center;
}

.critical-metric-large .stat-trend.good {
  color: var(--grafana-green);
  background-color: rgba(115, 191, 105, 0.15);
  border: 1px solid rgba(115, 191, 105, 0.3);
}

.critical-metric-large .stat-trend.warning {
  color: var(--grafana-orange);
  background-color: rgba(255, 183, 77, 0.15);
  border: 1px solid rgba(255, 183, 77, 0.3);
}

.critical-metric-large .stat-trend.critical {
  color: var(--grafana-red);
  background-color: rgba(242, 73, 92, 0.15);
  border: 1px solid rgba(242, 73, 92, 0.3);
}

/* Health Overview Section - Compact */
.health-overview-section {
  margin: 20px 0;
}

.health-overview-container {
  display: grid;
  grid-template-columns: auto 1fr auto;
  gap: 24px;
  align-items: center;
  background: linear-gradient(135deg, var(--grafana-panel-bg) 0%, rgba(82, 197, 247, 0.05) 100%);
  border: 2px solid var(--grafana-blue);
  border-radius: 8px;
  padding: 20px 24px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  position: relative;
}

.health-overview-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 4px;
  background: linear-gradient(90deg, var(--grafana-blue), var(--grafana-green));
  border-radius: 8px 8px 0 0;
}

/* Health Score Compact */
.health-score-compact {
  text-align: center;
  min-width: 120px;
}

.health-score-value-compact {
  font-size: 48px;
  font-weight: 900;
  line-height: 1;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  margin-bottom: 4px;
}

.health-score-value-compact.excellent {
  color: var(--grafana-green);
}

.health-score-value-compact.good {
  color: var(--grafana-blue);
}

.health-score-value-compact.warning {
  color: var(--grafana-orange);
}

.health-score-value-compact.critical {
  color: var(--grafana-red);
}

.health-score-label-compact {
  font-size: 11px;
  color: var(--grafana-text-muted);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: 600;
}

/* Secondary Stats Grid */
.secondary-stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  flex: 1;
  max-width: 600px;
}

.secondary-stat {
  text-align: center;
  padding: 12px;
  background: var(--grafana-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 6px;
  transition: background-color 0.2s ease;
}

.secondary-stat:hover {
  background: var(--grafana-bg-light);
}

.secondary-stat-value {
  font-size: 18px;
  font-weight: 700;
  color: var(--grafana-text);
  margin-bottom: 4px;
  line-height: 1;
}

.secondary-stat-label {
  font-size: 10px;
  color: var(--grafana-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 500;
}

/* Health Indicators Compact */
.health-indicators-compact {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 200px;
}

.health-indicator-compact {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  background-color: var(--grafana-bg-light);
  border-left: 3px solid var(--grafana-red);
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.health-indicator-compact:hover {
  background-color: var(--grafana-bg);
}

.health-indicator-compact.good {
  border-left-color: var(--grafana-green);
}

.health-indicator-compact.good i {
  color: var(--grafana-green);
}

.health-indicator-compact i {
  color: var(--grafana-red);
  width: 14px;
}

/* Secondary Metrics - Subtle */
.secondary-metric {
  background-color: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  opacity: 0.9;
}

/* Chart Panels - Tiered Hierarchy */
.chart-panel-primary {
  border: 2px solid var(--grafana-border);
  background: var(--grafana-panel-bg);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart-panel-primary .panel-header {
  background: linear-gradient(90deg, var(--grafana-bg-light), transparent);
  border-bottom: 2px solid var(--grafana-border);
}

.chart-panel-primary .panel-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--grafana-text);
}

.chart-panel-primary .panel-subtitle {
  font-size: 12px;
  color: var(--grafana-text-muted);
  margin-top: 4px;
  font-style: italic;
}

.chart-panel-secondary {
  border: 1px solid var(--grafana-border);
  background: var(--grafana-panel-bg);
}

.chart-panel-tertiary {
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: var(--grafana-panel-bg);
  opacity: 0.95;
}

/* Panel Spacing and Visual Hierarchy */
.panel {
  margin-bottom: 0;
}

/* Chart Container Heights */
.chart-container.time-series-chart {
  height: 320px;
  margin-bottom: 10px;
}

.chart-container.medium-chart {
  height: 280px;
}

.chart-container.small-chart {
  height: 250px;
}

.chart-container.endpoint-chart {
  height: 300px;
}

/* Stat panels improvements */
.panel.stat-panel {
  min-height: 130px;
}

.panel.stat-panel .panel-content {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  padding: 20px 15px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  margin-bottom: 4px;
  line-height: 1.1;
}

.stat-unit {
  font-size: 11px;
  color: var(--grafana-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
  font-weight: 500;
}

/* Color variants */
.stat-red {
  color: var(--grafana-red);
}

.stat-green {
  color: var(--grafana-green);
}

.stat-blue {
  color: var(--grafana-blue);
}

.stat-orange {
  color: var(--grafana-orange);
}

.stat-purple {
  color: #b877d9;
}

/* Health Score Panel - Enhanced */
.health-score-container {
  text-align: center;
  padding: 25px 20px;
}

.health-score-value {
  font-size: 52px;
  font-weight: 800;
  margin-bottom: 8px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  line-height: 1;
}

.health-score-value.excellent {
  color: var(--grafana-green);
}

.health-score-value.good {
  color: var(--grafana-blue);
}

.health-score-value.warning {
  color: var(--grafana-orange);
}

.health-score-value.critical {
  color: var(--grafana-red);
}

.health-score-label {
  font-size: 13px;
  color: var(--grafana-text-muted);
  margin-bottom: 20px;
  text-transform: uppercase;
  letter-spacing: 1px;
  font-weight: 600;
}

.health-indicators {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.health-indicator {
  padding: 8px 12px;
  border-radius: 6px;
  background-color: var(--grafana-bg-light);
  border-left: 3px solid var(--grafana-red);
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.health-indicator:hover {
  background-color: var(--grafana-bg);
}

.health-indicator.good {
  border-left-color: var(--grafana-green);
}



/* Responsive adjustments */
@media (max-width: 1400px) {
  .dashboard-grid {
    gap: 18px;
  }
  
  .critical-metric-large {
    min-height: 140px;
  }
  
  .stat-value-large {
    font-size: 36px;
  }
  
  .secondary-stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }
}

@media (max-width: 1200px) {
  .dashboard-grid {
    gap: 15px;
  }
  
  .critical-metric-large {
    grid-column: span 6;
    min-height: 120px;
  }
  
  .stat-value-large {
    font-size: 32px;
  }
  
  .health-overview-container {
    grid-template-columns: 1fr;
    gap: 20px;
    text-align: center;
  }
  
  .secondary-stats-grid {
    grid-template-columns: repeat(4, 1fr);
    max-width: none;
  }
  
  .health-indicators-compact {
    flex-direction: row;
    justify-content: center;
    flex-wrap: wrap;
    min-width: auto;
  }
  
  .performance-metric {
    grid-column: span 6;
  }
  
  .volume-metric {
    grid-column: span 4;
  }
  
  .panel.col-4 {
    grid-column: span 6;
  }
  
  .chart-container.time-series-chart {
    height: 280px;
  }
}

@media (max-width: 992px) {
  .critical-metric-large {
    grid-column: span 12;
    min-height: 110px;
  }
  
  .stat-value-large {
    font-size: 28px;
  }
  
  .health-score-value-compact {
    font-size: 36px;
  }
  
  .secondary-stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .performance-metric,
  .volume-metric {
    grid-column: span 6;
  }
  
  .panel.col-6 {
    grid-column: span 12;
  }
}

@media (max-width: 768px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
    gap: 15px;
  }
  
  .panel.col-12,
  .panel.col-8,
  .panel.col-6,
  .panel.col-4,
  .panel.col-3,
  .panel.col-2,
  .critical-metric-large,
  .performance-metric,
  .volume-metric,
  .health-overview-section {
    grid-column: span 1;
  }
  
  .critical-metric-large {
    min-height: 100px;
  }
  
  .stat-value-large {
    font-size: 24px;
  }
  
  .critical-metric-large .panel-content {
    padding: 16px;
  }
  
  .health-overview-container {
    padding: 16px;
    gap: 16px;
  }
  
  .health-score-value-compact {
    font-size: 32px;
  }
  
  .secondary-stats-grid {
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  
  .secondary-stat {
    padding: 8px;
  }
  
  .secondary-stat-value {
    font-size: 16px;
  }
  
  .health-indicators-compact {
    flex-direction: column;
    align-items: stretch;
  }
  
  .health-indicator-compact {
    justify-content: center;
  }
  
  .performance-metric .stat-value,
  .volume-metric .stat-value {
    font-size: 20px;
  }
  
  .volume-metric .stat-value {
    font-size: 18px;
  }
  
  .chart-container.time-series-chart,
  .chart-container.medium-chart,
  .chart-container.endpoint-chart {
    height: 250px;
  }
  
  .chart-container.small-chart {
    height: 220px;
  }
}

@media (max-width: 480px) {
  .critical-metric-large .stat-value-large {
    font-size: 20px;
  }
  
  .health-score-value-compact {
    font-size: 28px;
  }
  
  .secondary-stats-grid {
    grid-template-columns: 1fr;
  }
  
  .critical-metric-large .panel-content {
    padding: 12px;
  }
  
  .health-overview-container {
    padding: 12px;
  }
  
  .performance-metric .panel-content,
  .volume-metric .panel-content {
    padding: 12px;
  }
  
  .performance-metric .stat-value,
  .volume-metric .stat-value {
    font-size: 16px;
  }
}

/* Date Range Inputs and Export Button */
.date-range-container {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.analytics-date-input {
  padding: 8px 12px;
  background-color: var(--grafana-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 2px;
  color: var(--grafana-text);
  font-size: 14px;
  min-width: 140px;
  transition: border-color 0.2s ease;
}

.analytics-date-input:focus {
  outline: none;
  border-color: var(--grafana-blue);
}

.date-separator {
  color: var(--grafana-text-muted);
  font-size: 12px;
  font-weight: 500;
  margin: 0 4px;
}

.export-pdf-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: linear-gradient(135deg, var(--grafana-blue), #4a9eff);
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  min-width: 160px;
  justify-content: center;
}

.export-pdf-btn:hover {
  background: linear-gradient(135deg, #4a9eff, var(--grafana-blue));
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(74, 158, 255, 0.3);
}

.export-pdf-btn:active {
  transform: translateY(0);
}

.export-pdf-btn:disabled {
  background: var(--grafana-bg-light);
  color: var(--grafana-text-muted);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.export-pdf-btn:disabled:hover {
  background: var(--grafana-bg-light);
  transform: none;
  box-shadow: none;
}

.export-pdf-btn i {
  font-size: 14px;
}

/* Responsive adjustments for export controls */
@media (max-width: 768px) {
  .date-range-container {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .analytics-date-input {
    min-width: 100%;
  }
  
  .date-separator {
    text-align: center;
    margin: 4px 0;
  }
  
  .export-pdf-btn {
    width: 100%;
    min-width: auto;
  }
}
</style>
