<template>
  <div>
    <div v-if="loading" class="loading-state">Loading...</div>
    <div v-else>
      <header class="page-heading">
        <div>
          <p class="dashboard-eyebrow">STUDENT PORTAL</p>
          <h1>Payment history</h1>
          <p class="dashboard-subtitle">Track MoPay payment statuses and access receipts for confirmed payments.</p>
        </div>
      </header>

      <section v-if="selectedPayment" class="receipt-card">
      <div class="receipt-top">
        <span class="secure-note">● PAYMENT CONFIRMED</span>
        <span v-html="statusBadge(selectedPayment.status)"></span>
      </div>
      <h2>Payment receipt</h2>
      <p class="receipt-amount">{{ formatRWF(selectedPayment.amount) }}</p>
      <div class="receipt-grid">
        <div>
          <span>Receipt number</span>
          <strong>{{ selectedReceipt?.receipt_number || selectedPayment.receipt_number || selectedPayment.receipt || '—' }}</strong>
        </div>
        <div>
          <span>Invoice</span>
          <strong>{{ selectedPayment.invoice_number || selectedPayment.invoice_id }}</strong>
        </div>
        <div>
          <span>MoPay reference</span>
          <strong>{{ selectedPayment.mopay_reference || selectedPayment.provider_ref || '—' }}</strong>
        </div>
        <div>
          <span>Paid on</span>
          <strong>{{ formatDate(selectedPayment.status_updated_at || selectedPayment.created_at, { day: '2-digit', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' }) }}</strong>
        </div>
      </div>
      <div class="receipt-actions">
        <button 
          v-if="selectedReceipt"
          class="button button-primary" 
          @click="downloadReceiptPDF"
        >
          Download PDF
        </button>
        <button class="button button-secondary" @click="printReceipt">Print receipt</button>
      </div>
    </section>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>Transactions</h2>
          <p>Only payments confirmed as Paid have downloadable receipts.</p>
        </div>
        <span class="secure-note">● MoPay protected</span>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>PAYMENT</th>
              <th>INVOICE</th>
              <th>DATE</th>
              <th>AMOUNT</th>
              <th>STATUS</th>
              <th>RECEIPT</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="payment in payments" :key="payment.id">
              <td>
                <strong>{{ payment.id }}</strong>
                <small>{{ payment.provider || 'MoPay' }} · {{ payment.mopay_reference || '—' }}</small>
              </td>
              <td>{{ payment.invoice_number || payment.invoice_id }}</td>
              <td>{{ formatDate(payment.created_at) }}</td>
              <td><strong>{{ formatRWF(payment.amount) }}</strong></td>
              <td v-html="statusBadge(payment.status)"></td>
              <td>
                <button
                  v-if="payment.status === 'Paid'"
                  class="button button-quiet button-small"
                  @click="viewReceipt(payment)"
                >
                  View receipt
                </button>
                <span v-else>—</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="panel-footnote">The invoice balance changes only after MoPay confirms a payment as Paid.</p>
    </section>
    </div>
  </div>
</template>


<script setup>
import { ref, onMounted } from 'vue'
import { useApi, formatRWF, formatDate, statusBadge } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const api = useApi()
const payments = ref([])
const selectedPayment = ref(null)
const selectedReceipt = ref(null)
const loading = ref(true)

const viewReceipt = async (payment) => {
  console.log('viewReceipt called with payment:', payment)
  selectedPayment.value = payment
  selectedReceipt.value = null
  
  // Try to fetch receipt when viewing
  try {
    console.log('Fetching receipt for payment ID:', payment.id)
    const receipt = await api.getReceiptByPaymentId(payment.id)
    console.log('Receipt fetch result:', receipt)
    if (receipt && (receipt.id || receipt.receipt_number)) {
      selectedReceipt.value = receipt
      console.log('Receipt selected:', selectedReceipt.value)
    } else {
      console.log('No receipt data found in response')
    }
  } catch (error) {
    console.log('Receipt not available yet:', error)
  }
}

const downloadReceiptPDF = async () => {
  try {
    if (!selectedReceipt.value) {
      alert('Receipt not available yet. Please try again later.')
      return
    }
    
    const receipt = selectedReceipt.value
    const receiptId = receipt.id || receipt.receipt_number
    
    console.log('Downloading PDF for receipt ID:', receiptId)
    
    // Download PDF using receipt ID
    const blob = await api.getReceiptPDF(receiptId)
    console.log('Blob received, size:', blob.size)
    
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `receipt_${receipt.receipt_number || receiptId}.pdf`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    console.log('Download triggered')
  } catch (error) {
    console.error('Failed to download receipt:', error)
    alert('Failed to download receipt: ' + error.message)
  }
}

const printReceipt = () => {
  if (typeof window !== 'undefined') {
    window.print()
  }
}

const loadPayments = async () => {
  try {
    loading.value = true
    payments.value = await api.getPayments()
    console.log('Payments loaded:', payments.value)
    
    // Check if we should auto-select a payment from invoice navigation
    const selectedInvoiceId = localStorage.getItem('selectedInvoiceId')
    console.log('Looking for payment for invoice:', selectedInvoiceId)
    if (selectedInvoiceId) {
      // Match by invoice_number first, then invoice_id
      const relatedPayment = payments.value.find(p => 
        p.invoice_number === selectedInvoiceId || 
        String(p.invoice_id) === String(selectedInvoiceId)
      )
      console.log('Found payment:', relatedPayment)
      if (relatedPayment && relatedPayment.status === 'Paid') {
        await viewReceipt(relatedPayment)
      } else {
        console.log('No paid payment found for invoice:', selectedInvoiceId)
      }
      localStorage.removeItem('selectedInvoiceId')
    }
  } catch (error) {
    console.error('Failed to load payments:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadPayments()
})
</script>
