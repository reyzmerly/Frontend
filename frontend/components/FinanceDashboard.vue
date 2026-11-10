<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">FINANCE PORTAL</p>
        <h1>Finance overview</h1>
        <p class="dashboard-subtitle">Monitor invoices, collections and MoPay transactions across the university.</p>
      </div>
    </header>

    <section class="metric-grid metric-grid-4">
      <article class="metric-card">
        <span class="metric-icon icon-burgundy">▤</span>
        <p>Total invoiced</p>
        <strong>{{ formatRWF(issued) }}</strong>
        <small>{{ invoices.length }} invoices issued</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-green">✓</span>
        <p>Collected</p>
        <strong>{{ formatRWF(collected) }}</strong>
        <small>Confirmed by MoPay</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-gold">◷</span>
        <p>Outstanding</p>
        <strong>{{ formatRWF(outstanding) }}</strong>
        <small>{{ outstandingCount }} invoices to settle</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-slate">♙</span>
        <p>Students</p>
        <strong>{{ students.length }}</strong>
        <small>Active academic records</small>
      </article>
    </section>

    <div class="dashboard-grid-two">
      <section class="panel-card">
        <div class="section-heading">
          <div>
            <h2>Collections overview</h2>
            <p>Confirmed MoPay payments by month</p>
          </div>
          <span class="chart-legend"><i></i> Collected</span>
        </div>
        <div class="bar-chart" role="img" aria-label="Monthly confirmed collections">
          <div v-for="(value, index) in monthlyValues" :key="index" class="bar-column">
            <strong>{{ value ? formatRWF(value) : '—' }}</strong>
            <div class="bar-track">
              <span :style="{ height: Math.max(value / maxValue * 100, 4) + '%' }"></span>
            </div>
            <small>{{ months[index] }}</small>
          </div>
        </div>
      </section>

      <section class="panel-card">
        <div class="section-heading">
          <div>
            <h2>Invoice status</h2>
            <p>Current academic year</p>
          </div>
        </div>
        <div class="status-breakdown">
          <div v-for="status in statusTypes" :key="status">
            <span><small>{{ status }}</small></span>
            <span v-html="statusBadge(status)"></span>
            <strong>{{ getStatusCount(status) }}</strong>
          </div>
        </div>
        <div class="panel-divider"></div>
        <p class="panel-footnote">Payment statuses are supplied by MoPay. Finance staff cannot manually mark a payment as paid.</p>
      </section>
    </div>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>Recent MoPay payments</h2>
          <p>Latest gateway activity</p>
        </div>
        <div class="action-buttons">
          <button class="button button-secondary" @click="exportReport('pdf')">
            Export PDF
          </button>
          <button class="button button-secondary" @click="exportReport('csv')">
            Export CSV
          </button>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>STUDENT</th>
              <th>INVOICE</th>
              <th>MOPAY REFERENCE</th>
              <th>DATE</th>
              <th>AMOUNT</th>
              <th>STATUS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="payment in recentPayments" :key="payment.id">
              <td>
                <strong>{{ payment.student?.full_name || payment.student_name || '—' }}</strong>
                <small>{{ payment.student?.student_number || payment.student_id || '—' }}</small>
              </td>
              <td>{{ payment.invoice_number || payment.invoice_id }}</td>
              <td>{{ payment.mopay_reference || payment.provider_ref || '—' }}</td>
              <td>{{ formatDate(payment.created_at) }}</td>
              <td>{{ formatRWF(payment.amount) }}</td>
              <td v-html="statusBadge(payment.status)"></td>
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
const invoices = ref([])
const payments = ref([])
const students = ref([])
const loading = ref(true)

const months = ['Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov']
const statusTypes = ['Paid', 'Partially Paid', 'Draft', 'Issued', 'Overdue']

const issued = computed(() => {
  return invoices.value.reduce((sum, invoice) => sum + invoice.total_amount, 0)
})

const collected = computed(() => {
  return payments.value
    .filter(payment => payment.status === 'Paid')
    .reduce((sum, payment) => sum + payment.amount, 0)
})

const outstanding = computed(() => {
  return invoices.value.reduce((sum, invoice) => sum + (invoice.balance || 0), 0)
})

const outstandingCount = computed(() => {
  return invoices.value.filter(invoice => invoice.total_amount > invoice.paid_amount).length
})

const monthlyValues = computed(() => {
  return months.map((_, index) => {
    return payments.value
      .filter(payment => {
        return payment.status === 'Paid' && new Date(payment.created_at).getMonth() === index + 5
      })
      .reduce((sum, payment) => sum + payment.amount, 0)
  })
})

const maxValue = computed(() => {
  return Math.max(...monthlyValues.value, 1)
})

const recentPayments = computed(() => {
  return [...payments.value]
    .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
    .slice(0, 5)
})

const getStatusCount = (status) => {
  return invoices.value.filter(invoice => invoice.status === status).length
}

const exportReport = async (format) => {
  try {
    const blob = await api.exportFinanceReport(format)
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `financial-report.${format}`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  } catch (error) {
    console.error('Failed to export report:', error)
    alert('Failed to export report: ' + error.message)
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const [invoicesData, paymentsData, studentsData] = await Promise.all([
      api.getInvoices(),
      api.getPayments(),
      api.getStudents()
    ])
    invoices.value = invoicesData || []
    payments.value = paymentsData || []
    students.value = studentsData || []
  } catch (error) {
    console.error('Failed to load finance dashboard data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
