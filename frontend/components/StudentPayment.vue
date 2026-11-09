<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">STUDENT PORTAL / MY INVOICES</p>
        <h1>Pay invoice</h1>
        <p class="dashboard-subtitle">Make a full or partial payment through MoPay. Approve it on your phone.</p>
      </div>
    </header>

    <div v-if="!invoice">
      <p>Invoice not found</p>
      <button class="button button-secondary" @click="navigateToInvoices">Back to invoices</button>
    </div>

    <div v-else class="payment-layout">
      <section class="panel-card payment-form-card">
        <div class="section-heading">
          <div>
            <h2>Pay with MoPay</h2>
            <p>Choose how much you would like to pay.</p>
          </div>
          <span class="secure-note">● MoPay protected</span>
        </div>
        <div class="invoice-recap">
          <div><span>Invoice</span><strong>{{ invoice.id }}</strong></div>
          <div><span>Description</span><strong>{{ invoice.description }}</strong></div>
          <div><span>Outstanding balance</span><strong>{{ formatRWF(balance) }}</strong></div>
        </div>
        <form @submit.prevent="handleSubmit" class="stacked-form">
          <label for="payment-amount">Payment amount <span>(RWF)</span></label>
          <input
            id="payment-amount"
            v-model.number="paymentForm.amount"
            type="number"
            min="1"
            :max="balance"
            required
          >
          <small class="field-hint">Allowed: up to {{ formatRWF(balance) }}. Partial payments leave a balance.</small>
          <label for="mopay-phone">MoPay mobile number</label>
          <input
            id="mopay-phone"
            v-model="paymentForm.phone"
            type="tel"
            inputmode="tel"
            placeholder="+250 78 000 0142"
            maxlength="18"
            required
          >
          <div class="notice-box">
            <strong>Confirm only on your phone.</strong>
            <p>Any applicable gateway fees will be shown by MoPay before approval. Never share your PIN.</p>
          </div>
          <button class="button button-primary button-full" type="submit">
            Continue to MoPay <span>→</span>
          </button>
        </form>
      </section>

      <aside class="panel-card payment-summary">
        <p class="dashboard-eyebrow">PAYMENT SUMMARY</p>
        <h2>{{ formatRWF(balance) }}</h2>
        <p>Outstanding on this invoice</p>
        <hr>
        <div><span>Invoice total</span><strong>{{ formatRWF(invoice.total_amount) }}</strong></div>
        <div><span>Paid so far</span><strong>{{ formatRWF(invoice.paid_amount) }}</strong></div>
        <div><span>Due date</span><strong>{{ formatDate(invoice.due_date) }}</strong></div>
      </aside>
    </div>
  </div>
</template>


<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi, formatRWF, formatDate } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const invoice = ref(null)
const paymentForm = ref({
  amount: 0,
  phone: ''
})

const balance = computed(() => {
  if (!invoice.value) return 0
  return invoice.value.total_amount - invoice.value.paid_amount
})

const navigateToInvoices = () => {
  emit('navigate', 'My Invoices')
}

const handleSubmit = async () => {
  try {
    const paymentData = {
      invoice_id: invoice.value.id,
      amount: paymentForm.value.amount,
      phone: paymentForm.value.phone
    }

    const payment = await api.createPayment(paymentData)

    // Navigate to payment status page
    emit('navigate', 'Payment Status')
    localStorage.setItem('currentPayment', JSON.stringify(payment))
  } catch (error) {
    console.error('Payment failed:', error)
    alert('Payment failed: ' + error.message)
  }
}

onMounted(() => {
  const selectedInvoice = localStorage.getItem('selectedInvoice')
  if (selectedInvoice) {
    invoice.value = JSON.parse(selectedInvoice)
    paymentForm.value.amount = balance.value
  }
})
</script>
