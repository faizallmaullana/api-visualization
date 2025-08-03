<template>
  <!-- Sidebar Navigation -->
  <div>
    <nav class="sidebar">
      <!-- Logo Section -->
      <div class="sidebar-logo" @click="navigate('dashboard')">
        <div class="logo-icon">
          <i class="fas fa-tachometer-alt"></i>
        </div>
        <span class="logo-text">Tanjung API</span>
      </div>
      
      <!-- Menu Items -->
      <div 
        v-for="item in menuItems" 
        :key="item.name" 
        class="sidebar-item" 
        :class="{ active: currentSection === item.section }" 
        @click="navigate(item.section)"
        style="cursor: pointer;"
      >
        <i :class="item.icon"></i>
        <span class="sidebar-text">{{ item.label }}</span>
      </div>
    </nav>
    <!-- Top Navigation -->
    <nav class="topnav">
      <div class="topnav-title">{{ title }}</div>
      <div class="topnav-actions">
        <div class="refresh-interval"><i class="fas fa-sync-alt"></i> {{ refreshCounter }}s</div>
        <button class="time-range"><i class="fas fa-clock"></i> {{ timeRangeLabel }}</button>
      </div>
    </nav>
  </div>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';

const router = useRouter();
const route = useRoute();

const menuItems = [
  { section: 'dashboard', label: 'Dashboard', icon: 'fas fa-tachometer-alt' },
  { section: 'errors', label: 'Error Analysis', icon: 'fas fa-exclamation-triangle' },
  { section: 'timeline', label: 'Requests', icon: 'fas fa-chart-area' },
  { section: 'routes', label: 'Request Time', icon: 'fas fa-list' },
];

const currentSection = ref('dashboard');
const title = ref('API Analytics Dashboard');
const refreshCounter = ref(60);
const timeRangeLabel = ref('Last 24h');

function navigate(section) {
  console.log('Navigating to:', section);
  router.push({ name: section });
}

function updateCurrentSection() {
  const routeName = route.name || 'dashboard';
  currentSection.value = routeName;
  const menuItem = menuItems.find(i => i.section === routeName);
  title.value = menuItem ? `${menuItem.label} - API Analytics` : 'API Analytics Dashboard';
}

// Watch for route changes
watch(route, () => {
  updateCurrentSection();
});

// Initialize on mount
onMounted(() => {
  updateCurrentSection();
  // Start refresh counter
  setInterval(() => {
    refreshCounter.value = refreshCounter.value > 0 ? refreshCounter.value - 1 : 60;
  }, 1000);
});

// ... could add refresh counter logic here
</script>

<style scoped>
/* include sidebar and topnav styles from original CSS */
</style>
