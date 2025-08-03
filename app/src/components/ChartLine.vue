<template>
  <div class="chart-container time-series-chart">
    <canvas ref="canvasRef"></canvas>
  </div>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import Chart from 'chart.js/auto';

const props = defineProps({ labels: Array, datasets: Array });
const canvasRef = ref(null);
let chartInstance;

onMounted(() => {
  const ctx = canvasRef.value.getContext('2d');
  chartInstance = new Chart(ctx, {
    type: 'line',
    data: { labels: props.labels, datasets: props.datasets },
    options: { responsive: true, maintainAspectRatio: false, scales: { x: { type: 'time' }, y: { beginAtZero: true }, y1: { position: 'right' } } }
  });
});

watch(() => props.labels, (newLabels) => {
  chartInstance.data.labels = newLabels;
  chartInstance.update();
});

watch(() => props.datasets, (newData) => {
  chartInstance.data.datasets = newData;
  chartInstance.update();
});
</script>

<style scoped>
/* Use .chart-container styles */
</style>
