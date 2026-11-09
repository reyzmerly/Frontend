<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">ADMIN PORTAL</p>
        <h1>Audit log</h1>
        <p class="dashboard-subtitle">Read-only history of account, invoice and payment activity.</p>
      </div>
      <button class="button button-secondary" @click="navigateTo('Admin overview')">
        ← Back to Dashboard
      </button>
    </header>

    <section class="notice-box notice-neutral">
      <strong>Read-only audit history</strong>
      <p>Audit entries are append-only in this demo view. Changes to payment outcomes are made by MoPay, not staff.</p>
    </section>

    <section class="panel-card">
      <div class="toolbar">
        <div>
          <h2>Activity records</h2>
          <p>{{ auditLog.length }} events</p>
        </div>
        <input
          v-model="searchTerm"
          type="search"
          placeholder="Search activity…"
          aria-label="Search audit log"
        >
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>DATE & TIME</th>
              <th>ACTOR TYPE</th>
              <th>ACTION</th>
              <th>RECORD REF</th>
              <th>TYPE</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="entry in filteredAuditLog" :key="entry.id">
              <td>{{ formatDate(entry.created_at, { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }) }}</td>
              <td><strong>{{ entry.actor_type || '—' }}</strong></td>
              <td>{{ entry.action }}</td>
              <td>{{ entry.record_ref || '—' }}</td>
              <td><span class="tag-pill">{{ entry.record_type }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>


<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi, formatDate } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const auditLog = ref([])
const loading = ref(true)
const searchTerm = ref('')

const navigateTo = (page) => {
  emit('navigate', page)
}

const filteredAuditLog = computed(() => {
  const term = searchTerm.value.toLowerCase()
  return auditLog.value.filter(entry => {
    return `${entry.actor_type} ${entry.action} ${entry.record_ref} ${entry.record_type}`.toLowerCase().includes(term)
  })
})

const loadAuditLog = async () => {
  try {
    auditLog.value = await api.getAuditLogs()
  } catch (error) {
    console.error('Failed to load audit log:', error)
  }
}

onMounted(() => {
  loadAuditLog()
})
</script>
