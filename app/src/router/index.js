import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../components/Dashboard.vue';
import Analytics from '../components/Analytics.vue';
import Routes from '../components/Routes.vue';
import Timeline from '../components/Timeline.vue';
import AnalyticsDashboard from '../components/AnalyticsDashboard.vue';
import ErrorAnalysis from '../components/ErrorAnalysis.vue';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'dashboard', component: Dashboard },
    { path: '/analytics', name: 'analytics', component: Analytics },
    { path: '/routes', name: 'routes', component: Routes },
    { path: '/timeline', name: 'timeline', component: Timeline },
    { path: '/history', name: 'history', component: AnalyticsDashboard },
    { path: '/errors', name: 'errors', component: ErrorAnalysis }
  ],
})

export default router
