<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">FINANCE PORTAL</p>
        <h1>Financial reports</h1>
        <p class="dashboard-subtitle">Review issued invoices and confirmed collections. Reports exclude pending and failed MoPay requests.</p>
      </div>
      <div class="page-heading-action">
        <button class="button button-secondary" @click="exportReport">↓ Export CSV</button>
      </div>
    </header>

    <section class="metric-grid metric-grid-3">
      <article class="metric-card">
        <span class="metric-icon icon-burgundy">▤</span>
        <p>Total invoiced</p>
        <strong>{{ formatRWF(totalInvoiced) }}</strong>
        <small>{{ invoices.length }} invoices</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-green">✓</span>
        <p>Collected</p>
        <strong>{{ formatRWF(collected) }}</strong>
        <small>{{ paid.length }} confirmed payments</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-gold">◷</span>
        <p>Outstanding</p>
        <strong>{{ formatRWF(outstanding) }}</strong>
        <small>Unpaid invoice balances</small>
      </article>
    </section>

    <div class="dashboard-grid-two">
      <section class="panel-card">
        <div class="section-heading">
          <div>
            <h2>Collections by fee</h2>
            <p>Recorded payment amounts by invoice type</p>
          </div>
        </div>
        <div v-for="(amount, name) in collectionByFee" :key="name" class="report-bar-row">
          <div>
            <strong>{{ name }}</strong>
            <span>{{ formatRWF(amount) }}</span>
          </div>
          <div class="progress-track">
            <i :style="{ width: (amount / collected * 100) + '%' }"></i>
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
        <p class="panel-footnote">Payment states and references are provided by MoPay. Staff cannot manually change transaction outcomes.</p>
      </section>
    </div>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>Confirmed collection records</h2>
          <p>Read-only gateway records</p>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>DATE</th>
              <th>STUDENT</th>
              <th>INVOICE</th>
              <th>REFERENCE</th>
              <th>AMOUNT</th>
              <th>STATUS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="payment in paid" :key="payment.id">
              <td>{{ formatDate(payment.created_at) }}</td>
              <td>{{ payment.student_name || '—' }}</td>
              <td>{{ payment.invoice_id }}</td>
              <td>{{ payment.provider_ref || '—' }}</td>
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
const loading = ref(true)

const statusTypes = ['Paid', 'Partially Paid', 'Draft', 'Issued', 'Overdue']

const paid = computed(() => {
  return payments.value.filter(payment => payment.status === 'Paid')
})

const collected = computed(() => {
  return paid.value.reduce((sum, payment) => sum + payment.amount, 0)
})

const totalInvoiced = computed(() => {
  return invoices.value.reduce((sum, invoice) => sum + invoice.total_amount, 0)
})

const outstanding = computed(() => {
  return invoices.value.reduce((sum, invoice) => sum + invoice.total_amount - invoice.paid_amount, 0)
})

const collectionByFee = computed(() => {
  return invoices.value.reduce((summary, invoice) => {
    const category = invoice.description.split('—')[0].trim()
    summary[category] = (summary[category] || 0) + invoice.paid_amount
    return summary
  }, {})
})

const getStatusCount = (status) => {
  return invoices.value.filter(invoice => invoice.status === status).length
}

const exportReport = async () => {
  try {
    const blob = await api.exportFinanceReport('csv')
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = 'uok-payment-report.csv'
    link.click()
    window.URL.revokeObjectURL(url)
  } catch (error) {
    console.error('Failed to export report:', error)
    alert('Failed to export report: ' + error.message)
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const [invoicesData, paymentsData] = await Promise.all([
      api.getInvoices(),
      api.getPayments()
    ])
    invoices.value = invoicesData || []
    payments.value = paymentsData || []
  } catch (error) {
    console.error('Failed to load finance reports data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
