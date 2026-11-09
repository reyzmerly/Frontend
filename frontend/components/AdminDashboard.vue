<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">ADMIN PORTAL</p>
        <h1>Admin overview</h1>
        <p class="dashboard-subtitle">Manage system-wide access and view audit logs.</p>
      </div>
    </header>

    <section class="metric-grid metric-grid-3">
      <article class="metric-card">
        <span class="metric-icon icon-burgundy">▤</span>
        <p>Total users</p>
        <strong>{{ users.length }}</strong>
        <small>Registered accounts</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-green">✓</span>
        <p>Active accounts</p>
        <strong>{{ activeUsers }}</strong>
        <small>Currently active</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-gold">◷</span>
        <p>Audit events</p>
        <strong>{{ auditLog.length }}</strong>
        <small>Recorded activities</small>
      </article>
    </section>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>Quick actions</h2>
          <p>Common administrative tasks</p>
        </div>
      </div>
      <div class="action-grid">
        <button class="button button-secondary" @click="navigateTo('Roles and access')">
          Manage user roles
        </button>
        <button class="button button-secondary" @click="navigateTo('Audit Log')">
          View audit log
        </button>
      </div>
    </section>
  </div>
</template>


<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const users = ref([])
const auditLog = ref([])
const loading = ref(true)

const activeUsers = computed(() => {
  return users.value.filter(user => user.status === 'Active').length
})

const navigateTo = (page) => {
  emit('navigate', page)
}

const loadData = async () => {
  try {
    loading.value = true
    const [usersData, auditData] = await Promise.all([
      api.getUsers(),
      api.getAuditLogs()
    ])
    users.value = usersData || []
    auditLog.value = auditData || []
  } catch (error) {
    console.error('Failed to load admin dashboard data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
