<template>
  <div class="page-header">
    <!-- Page Title Section -->
    <div class="page-header-main">
      <div class="page-title-section">
        <div class="page-icon">
          <i :class="pageIcon"></i>
        </div>
        <div class="page-info">
          <h1 class="page-title">{{ pageTitle }}</h1>
          <p class="page-description">{{ pageDescription }}</p>
        </div>
      </div>
      
      <!-- Page Stats (if provided) -->
      <div v-if="stats && stats.length" class="page-stats">
        <div 
          v-for="stat in stats" 
          :key="stat.label" 
          class="stat-item"
          :class="stat.type"
        >
          <div class="stat-value">{{ stat.value }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
      </div>

      <!-- Page Actions -->
      <div class="page-actions">
        <div class="page-meta">
          <div class="last-updated">
            <i class="fas fa-clock"></i>
            <span>Updated {{ lastUpdated }}</span>
          </div>
          <div class="auto-refresh" :class="{ active: autoRefresh }">
            <i class="fas fa-sync-alt" :class="{ spinning: refreshing }"></i>
            <span>{{ refreshCounter }}s</span>
          </div>
        </div>
        <div class="action-buttons">
          <button 
            class="action-btn refresh-btn" 
            @click="$emit('refresh')"
            :disabled="refreshing"
          >
            <i class="fas fa-sync-alt" :class="{ spinning: refreshing }"></i>
            Refresh
          </button>
          <button 
            v-if="showExport" 
            class="action-btn export-btn" 
            @click="$emit('export')"
          >
            <i class="fas fa-download"></i>
            Export
          </button>
        </div>
      </div>
    </div>

    <!-- Filters Section -->
    <div v-if="$slots.filters" class="page-filters">
      <div class="filters-container">
        <slot name="filters"></slot>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';

const props = defineProps({
  pageTitle: {
    type: String,
    required: true
  },
  pageDescription: {
    type: String,
    default: ''
  },
  pageIcon: {
    type: String,
    default: 'fas fa-chart-line'
  },
  stats: {
    type: Array,
    default: () => []
  },
  showExport: {
    type: Boolean,
    default: false
  },
  autoRefresh: {
    type: Boolean,
    default: true
  },
  refreshInterval: {
    type: Number,
    default: 60
  },
  refreshing: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['refresh', 'export']);

const refreshCounter = ref(props.refreshInterval);
const lastUpdated = ref('');
let refreshTimer = null;

const updateLastUpdated = () => {
  const now = new Date();
  lastUpdated.value = now.toLocaleTimeString();
};

const startRefreshTimer = () => {
  if (!props.autoRefresh) return;
  
  refreshTimer = setInterval(() => {
    refreshCounter.value--;
    if (refreshCounter.value <= 0) {
      refreshCounter.value = props.refreshInterval;
      emit('refresh');
    }
  }, 1000);
};

const stopRefreshTimer = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
};

const resetRefreshCounter = () => {
  refreshCounter.value = props.refreshInterval;
  updateLastUpdated();
};

onMounted(() => {
  updateLastUpdated();
  startRefreshTimer();
});

onUnmounted(() => {
  stopRefreshTimer();
});

// Reset counter when refresh completes
const handleRefresh = () => {
  resetRefreshCounter();
  emit('refresh');
};

defineExpose({
  resetRefreshCounter
});
</script>

<style scoped>
.page-header {
  background: var(--grafana-panel-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 4px;
  margin-bottom: 24px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.page-header-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--grafana-border);
}

.page-title-section {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.page-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, var(--grafana-blue), var(--grafana-green));
  border-radius: 8px;
  color: white;
  font-size: 20px;
}

.page-info h1.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--grafana-text);
  margin: 0 0 4px 0;
  line-height: 1.2;
}

.page-info .page-description {
  font-size: 14px;
  color: var(--grafana-text-muted);
  margin: 0;
  line-height: 1.4;
}

.page-stats {
  display: flex;
  gap: 24px;
  margin: 0 24px;
}

.stat-item {
  text-align: center;
  padding: 12px 16px;
  background: var(--grafana-bg);
  border: 1px solid var(--grafana-border);
  border-radius: 6px;
  min-width: 80px;
}

.stat-item.critical {
  border-left: 4px solid var(--grafana-red);
}

.stat-item.warning {
  border-left: 4px solid var(--grafana-orange);
}

.stat-item.success {
  border-left: 4px solid var(--grafana-green);
}

.stat-item.info {
  border-left: 4px solid var(--grafana-blue);
}

.stat-value {
  font-size: 18px;
  font-weight: 700;
  color: var(--grafana-text);
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 11px;
  color: var(--grafana-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 500;
}

.page-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
}

.page-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 12px;
  color: var(--grafana-text-muted);
}

.last-updated, .auto-refresh {
  display: flex;
  align-items: center;
  gap: 6px;
}

.auto-refresh.active {
  color: var(--grafana-green);
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border: 1px solid var(--grafana-border);
  border-radius: 4px;
  background: var(--grafana-bg);
  color: var(--grafana-text);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover:not(:disabled) {
  background: var(--grafana-bg-light);
  border-color: var(--grafana-blue);
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.refresh-btn {
  background: var(--grafana-blue);
  color: white;
  border-color: var(--grafana-blue);
}

.refresh-btn:hover:not(:disabled) {
  background: rgba(82, 197, 247, 0.8);
}

.export-btn {
  background: var(--grafana-green);
  color: white;
  border-color: var(--grafana-green);
}

.export-btn:hover {
  background: rgba(115, 191, 105, 0.8);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Filters Section */
.page-filters {
  padding: 16px 24px;
  background: var(--grafana-bg);
  border-bottom-left-radius: 4px;
  border-bottom-right-radius: 4px;
}

.filters-container {
  display: flex;
  gap: 16px;
  align-items: end;
  flex-wrap: wrap;
}

/* Responsive Design */
@media (max-width: 1200px) {
  .page-header-main {
    flex-direction: column;
    align-items: stretch;
    gap: 16px;
  }
  
  .page-title-section {
    justify-content: flex-start;
  }
  
  .page-stats {
    justify-content: center;
    margin: 0;
  }
  
  .page-actions {
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
}

@media (max-width: 768px) {
  .page-header-main {
    padding: 16px;
  }
  
  .page-title-section {
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 12px;
  }
  
  .page-stats {
    flex-wrap: wrap;
    gap: 12px;
  }
  
  .stat-item {
    min-width: auto;
    flex: 1;
  }
  
  .page-actions {
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }
  
  .action-buttons {
    flex-direction: column;
    width: 100%;
  }
  
  .action-btn {
    justify-content: center;
  }
  
  .page-filters {
    padding: 12px 16px;
  }
  
  .filters-container {
    flex-direction: column;
    gap: 12px;
  }
}
</style>
