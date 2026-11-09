<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">STUDENT PORTAL</p>
        <h1>My invoices</h1>
        <p class="dashboard-subtitle">Only your own invoices are shown. Receipts are available from Payment history for Paid payments.</p>
      </div>
    </header>

    <section class="metric-grid metric-grid-3">
      <article class="metric-card">
        <span class="metric-icon icon-burgundy">▤</span>
        <p>Invoices this semester</p>
        <strong>{{ invoices.length }}</strong>
        <small>Academic year 2026/27</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-gold">◷</span>
        <p>Amount due</p>
        <strong>{{ formatRWF(outstandingTotal) }}</strong>
        <small>{{ outstanding.length }} invoice{{ outstanding.length === 1 ? '' : 's' }} to settle</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-green">✓</span>
        <p>Paid invoices</p>
        <strong>{{ paidCount }}</strong>
        <small>Receipts in payment history</small>
      </article>
    </section>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>All invoices</h2>
          <p>Amounts in RWF · Partial payments leave a balance.</p>
        </div>
        <span class="secure-note">● MoPay protected</span>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>INVOICE</th>
              <th>FEE DESCRIPTION</th>
              <th>AMOUNT</th>
              <th>BALANCE</th>
              <th>DUE DATE</th>
              <th>STATUS</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="invoice in invoices" :key="invoice.id" class="clickable-row" @click="showInvoiceDetails(invoice)">
              <td>
                <strong>{{ invoice.invoice_number || invoice.id }}</strong>
                <small>Issued {{ formatDate(invoice.created_at) }}</small>
              </td>
              <td>{{ invoice.description }}</td>
              <td>{{ formatRWF(invoice.total_amount) }}</td>
              <td><strong>{{ formatRWF(invoice.balance) }}</strong></td>
              <td>{{ formatDate(invoice.due_date) }}</td>
              <td v-html="statusBadge(invoice.status)"></td>
              <td @click.stop>
                <button
                  v-if="invoice.balance > 0"
                  class="button button-primary button-small"
                  @click="initiatePayment(invoice)"
                >
                  Pay now
                </button>
                <button
                  v-else
                  class="button button-quiet button-small"
                  @click="navigateToHistory(invoice)"
                >
                  Receipt
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="panel-footnote">Payments are confirmed on your phone. Never share your MoPay PIN. Click on a row to view details.</p>
    </section>

    <InvoiceDetails
      v-if="selectedInvoice"
      :invoice="selectedInvoice"
      :user="user"
      @close="selectedInvoice = null"
      @navigate="handleNavigate"
    />
  </div>
</template>


<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi, formatRWF, formatDate, statusBadge } from '~/composables/useApi'
import InvoiceDetails from './InvoiceDetails.vue'

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
const selectedInvoice = ref(null)
const loading = ref(true)

const outstanding = computed(() => {
  return invoices.value.filter(invoice => invoice.balance > 0)
})

const outstandingTotal = computed(() => {
  return outstanding.value.reduce((sum, invoice) => sum + invoice.balance, 0)
})

const paidCount = computed(() => {
  return invoices.value.filter(invoice => invoice.status === 'Paid').length
})

const showInvoiceDetails = async (invoice) => {
  console.log('Showing invoice details for:', invoice)
  try {
    const fullInvoice = await api.getInvoiceById(invoice.id)
    console.log('Full invoice from API:', fullInvoice)
    selectedInvoice.value = fullInvoice
    console.log('selectedInvoice set to:', selectedInvoice.value)
  } catch (error) {
    console.error('Failed to load invoice details:', error)
    alert('Failed to load invoice details: ' + error.message)
  }
}

const handleNavigate = (page) => {
  emit('navigate', page)
}

const initiatePayment = (invoice) => {
  emit('navigate', 'Payment')
  localStorage.setItem('selectedInvoice', JSON.stringify(invoice))
}

const navigateToHistory = (invoice) => {
  console.log('Receipt button clicked for invoice:', invoice)
  const invoiceId = invoice.invoice_number || invoice.id
  console.log('Storing invoice ID:', invoiceId)
  localStorage.setItem('selectedInvoiceId', invoiceId)
  console.log('Navigating to Payment History')
  emit('navigate', 'Payment History')
}

const loadInvoices = async () => {
  try {
    loading.value = true
    invoices.value = await api.getInvoices()
  } catch (error) {
    console.error('Failed to load invoices:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadInvoices()
})
</script>

<style scoped>
.clickable-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.clickable-row:hover {
  background-color: #f9fafb;
}
</style>
