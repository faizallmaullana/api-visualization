<template>
  <div class="analytics-filters">
    <!-- Filter controls: Time Range, Endpoint, Method -->
    <div class="filter-group">
      <label class="filter-label">Time Range</label>
      <select class="filter-select" v-model="timeRange">
        <option value="1h">Last 1 hour</option>
        <option value="6h">Last 6 hours</option>
        <option value="24h">Last 24 hours</option>
        <option value="7d">Last 7 days</option>
      </select>
    </div>
    <div class="filter-group">
      <label class="filter-label">Endpoint</label>
      <select class="filter-select" v-model="endpoint">
        <option value="all">All Endpoints</option>
      </select>
    </div>
    <div class="filter-group">
      <label class="filter-label">Method</label>
      <select class="filter-select" v-model="method">
        <option value="all">All Methods</option>
        <option>GET</option>
        <option>POST</option>
        <option>PUT</option>
        <option>DELETE</option>
      </select>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, defineEmits } from 'vue';
const emit = defineEmits(['filter-change']);
const timeRange = ref('1h');
const endpoint = ref('all');
const method = ref('all');

watch([timeRange, endpoint, method], () => {
  emit('filter-change', { timeRange: timeRange.value, endpoint: endpoint.value, method: method.value });
});
</script>

<style scoped>
/* Copy styles from original CSS for .analytics-filters, .filter-group, .filter-label, .filter-select */
.filter-group { display: flex; flex-direction: column; gap: 5px; }
.filter-label { font-size: 12px; color: #8e8e8e; }
.filter-select { background: #0f1419; border: 1px solid #262626; color: #d9d9d9; padding: 6px 10px; border-radius: 3px; font-size: 13px; }
.analytics-filters { background: #1e2028; border: 1px solid #262626; border-radius: 4px; padding: 15px; margin-bottom: 20px; display: flex; gap: 15px; align-items: center; }
</style>
