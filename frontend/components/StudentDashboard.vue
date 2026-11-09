<template>
  <div>
    <div v-if="loading" class="loading-state">Loading...</div>
    <div v-else>
      <section class="dashboard-heading">
        <div>
          <p class="dashboard-eyebrow">STUDENT PORTAL</p>
          <h1>Welcome back, {{ user.full_name }}</h1>
          <p class="dashboard-subtitle">Here is an overview of your university financial records.</p>
        </div>
        <div class="academic-context"><span>2026/27</span><strong>Semester 1</strong></div>
      </section>

      <section class="student-profile-card">
        <div class="student-avatar">{{ initials }}</div>
        <div>
          <strong>{{ user.full_name }}</strong>
          <p>{{ user.email }}</p>
          <small>Student</small>
        </div>
      </section>

      <section class="metric-grid metric-grid-3">
        <article class="metric-card">
          <span class="metric-icon icon-burgundy">▤</span>
          <p>Total invoiced</p>
          <strong>{{ formatRWF(summary.totalFees) }}</strong>
          <small>{{ invoices.length }} invoices this semester</small>
        </article>
        <article class="metric-card">
          <span class="metric-icon icon-green">✓</span>
          <p>Amount paid</p>
          <strong>{{ formatRWF(summary.amountPaid) }}</strong>
          <small>Confirmed by MoPay</small>
        </article>
        <article class="metric-card">
          <span class="metric-icon icon-gold">◷</span>
          <p>Outstanding balance</p>
          <strong>{{ formatRWF(summary.outstanding) }}</strong>
          <small>Amount remaining to pay</small>
        </article>
      </section>

      <section class="panel-card">
        <div class="section-heading">
          <div>
            <h2>Recent invoices</h2>
            <p>Only your own invoices are shown. Partial payments leave a balance.</p>
          </div>
          <button class="text-action" @click="emit('navigate', 'My Invoices')">View all invoices →</button>
        </div>
        <div class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th>INVOICE</th>
                <th>DESCRIPTION</th>
                <th>DUE DATE</th>
                <th>AMOUNT</th>
                <th>BALANCE</th>
                <th>STATUS</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="invoices.length === 0">
                <td colspan="7" class="empty-state">No invoices found.</td>
              </tr>
              <tr v-for="invoice in invoices.slice(0, 5)" :key="invoice.id">
                <td><strong>{{ invoice.invoice_number || invoice.id }}</strong></td>
                <td>{{ invoice.description }}</td>
                <td>{{ formatDate(invoice.due_date) }}</td>
                <td>{{ formatRWF(invoice.total_amount) }}</td>
                <td><strong>{{ formatRWF(invoice.balance) }}</strong></td>
                <td v-html="statusBadge(invoice.status)"></td>
                <td>
                  <button
                    v-if="invoice.balance > 0"
                    class="button button-primary button-small"
                    @click="initiatePayment(invoice)"
                  >
                    Pay now
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="panel-card">
        <div class="section-heading">
          <div>
            <h2>Recent transactions</h2>
            <p>Your latest MoPay activity</p>
          </div>
          <button class="text-action" @click="emit('navigate', 'Payment History')">Payment history →</button>
        </div>
        <div class="transaction-list">
          <div v-if="payments.length === 0" class="empty-state">No payment activity yet.</div>
          <div v-for="payment in payments.slice(0, 4)" :key="payment.id" class="transaction-item clickable" @click="viewPaymentReceipt(payment)">
            <div class="transaction-icon">↗</div>
            <div class="transaction-description">
              <strong>{{ getInvoiceDescription(payment.invoice_id) }}</strong>
              <small>{{ payment.id }} · {{ formatDate(payment.created_at) }}</small>
            </div>
            <div class="transaction-right">
              <strong>{{ formatRWF(payment.amount) }}</strong>
              <span v-html="statusBadge(payment.status)"></span>
            </div>
          </div>
        </div>
        <p class="panel-footnote">Payments are confirmed on your phone. Never share your MoPay PIN.</p>
      </section>
    </div>
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

const initials = computed(() => {
  if (!props.user) return ''
  return props.user.full_name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
})

const summary = computed(() => {
  return {
    totalFees: invoices.value.reduce((sum, invoice) => sum + (invoice.total_amount || 0), 0),
    amountPaid: invoices.value.reduce((sum, invoice) => sum + (invoice.paid_amount || 0), 0),
    outstanding: invoices.value.reduce((sum, invoice) => sum + (invoice.balance || 0), 0)
  }
})

const getInvoiceDescription = (invoiceId) => {
  const invoice = invoices.value.find(i => i.id === invoiceId || i.invoice_number === invoiceId)
  return invoice ? invoice.description : 'University payment'
}

const initiatePayment = (invoice) => {
  // Navigate to payment page with invoice ID
  emit('navigate', 'Payment')
  // Store selected invoice
  localStorage.setItem('selectedInvoice', JSON.stringify(invoice))
}

const viewPaymentReceipt = (payment) => {
  if (payment.status === 'Paid') {
    localStorage.setItem('selectedInvoiceId', payment.invoice_id)
    emit('navigate', 'Payment History')
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const invoiceData = await api.getInvoices()
    const paymentData = await api.getPayments()
    console.log('Invoices received:', invoiceData)
    console.log('Payments received:', paymentData)
    invoices.value = invoiceData || []
    payments.value = paymentData || []
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
