<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">FINANCE PORTAL</p>
        <h1>Payments monitor</h1>
        <p class="dashboard-subtitle">Monitor MoPay transaction records and add follow-up notes. Status changes are read-only.</p>
      </div>
    </header>

    <section class="metric-grid metric-grid-3">
      <article class="metric-card">
        <span class="metric-icon icon-green">✓</span>
        <p>Confirmed today</p>
        <strong>{{ confirmedToday }}</strong>
        <small>Based on MoPay confirmation</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-gold">◷</span>
        <p>Awaiting confirmation</p>
        <strong>{{ pendingCount }}</strong>
        <small>Do not submit a duplicate request</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-burgundy">↗</span>
        <p>Confirmed volume</p>
        <strong>{{ formatRWF(paidTotal) }}</strong>
        <small>Paid transactions only</small>
      </article>
    </section>

    <section class="notice-box notice-neutral">
      <strong>VIEW ONLY · Records and statuses come from MoPay.</strong>
      <p>Staff can filter records and add follow-up notes only. Do not manually mark a payment as paid or failed.</p>
    </section>

    <section class="panel-card">
      <div class="toolbar">
        <div>
          <h2>Gateway transactions</h2>
          <p>{{ payments.length }} MoPay records</p>
        </div>
        <div class="toolbar-controls">
          <input
            v-model="searchTerm"
            type="search"
            placeholder="Student, invoice or reference…"
            aria-label="Search payments"
          >
          <select v-model="statusFilter" aria-label="Filter by payment status">
            <option value="">All statuses</option>
            <option>Paid</option>
            <option>Pending</option>
            <option>Failed</option>
            <option>Cancelled</option>
          </select>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>DATE</th>
              <th>STUDENT</th>
              <th>INVOICE</th>
              <th>MOPAY REFERENCE</th>
              <th>AMOUNT</th>
              <th>STATUS</th>
              <th>FOLLOW-UP</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="payment in filteredPayments" :key="payment.id" :data-status="payment.status">
              <td>
                {{ formatDate(payment.created_at) }}
                <small>{{ formatTime(payment.created_at) }}</small>
              </td>
              <td>
                <strong>{{ payment.student_name || '—' }}</strong>
                <small>{{ payment.student_id || '—' }}</small>
              </td>
              <td>{{ payment.invoice_id }}</td>
              <td>{{ payment.provider_ref || '—' }}</td>
              <td>{{ formatRWF(payment.amount) }}</td>
              <td v-html="statusBadge(payment.status)"></td>
              <td>
                <button
                  class="button button-quiet button-small"
                  @click="addNote(payment)"
                >
                  {{ payment.note ? 'Edit note' : 'Add note' }}
                </button>
              </td>
            </tr>
            <tr v-if="payment.note" class="note-row">
              <td colspan="7">
                <strong>Follow-up:</strong> {{ payment.note }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>


<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi, formatRWF, formatDate, statusBadge } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const payments = ref([])
const loading = ref(true)

const searchTerm = ref('')
const statusFilter = ref('')

const confirmedToday = computed(() => {
  const today = new Date().toDateString()
  return payments.value.filter(payment => {
    return payment.status === 'Paid' && new Date(payment.created_at).toDateString() === today
  }).length
})

const pendingCount = computed(() => {
  return payments.value.filter(payment => payment.status === 'Pending').length
})

const paidTotal = computed(() => {
  return payments.value
    .filter(payment => payment.status === 'Paid')
    .reduce((sum, payment) => sum + payment.amount, 0)
})

const filteredPayments = computed(() => {
  const search = searchTerm.value.toLowerCase()
  const status = statusFilter.value

  return payments.value.filter(payment => {
    const matchesSearch = `${payment.student_name} ${payment.student_id} ${payment.invoice_id} ${payment.provider_ref}`.toLowerCase().includes(search)
    const matchesStatus = !status || payment.status === status
    return matchesSearch && matchesStatus
  })
})

const formatTime = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const addNote = (payment) => {
  const note = window.prompt('Add a follow-up note. This does not change the payment status:', payment.note || '')
  if (note === null) return

  // In a real app, you would call an API to update the note
  payment.note = note.trim()
  // Reload payments
  loadPayments()
}

const loadPayments = async () => {
  try {
    payments.value = await api.getPayments()
  } catch (error) {
    console.error('Failed to load payments:', error)
  }
}

onMounted(() => {
  loadPayments()
})
</script>
