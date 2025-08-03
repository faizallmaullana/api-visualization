<template>
  <div class="panel stat-panel col-3">
    <div class="panel-header">
      <div class="panel-title">{{ title }}</div>
    </div>
    <div class="panel-content">
      <div class="stat-value">{{ value }}</div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  title: { type: String, required: true },
  value: { type: [String, Number], required: true }
});
</script>

<style scoped>
/* Use global panel and stat-panel styles */
</style>
