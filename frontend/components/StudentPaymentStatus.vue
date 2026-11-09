<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">STUDENT PORTAL / PAYMENT HISTORY</p>
        <h1>{{ title }}</h1>
        <p class="dashboard-subtitle">{{ description }}</p>
      </div>
    </header>

    <section v-if="payment" :class="['status-panel', statusPanelClass]">
      <div class="status-orb">{{ statusOrb }}</div>
      <span v-html="statusBadge(payment.status)"></span>
      <h2>{{ formatRWF(payment.amount) }}</h2>
      <p>{{ statusMessage }}</p>
      <div class="reference-card">
        <span>MoPay reference</span>
        <strong>{{ payment.provider_ref || '—' }}</strong>
      </div>

      <div v-if="payment.status === 'Pending'" class="demo-actions">
        <p>Demo preview — no real payment will be sent.</p>
        <button class="button button-primary" @click="simulatePaid">Simulate MoPay success</button>
        <button class="button button-secondary" @click="simulateFailed">Simulate failed payment</button>
        <button class="text-action" @click="cancelPayment">Cancel payment</button>
      </div>

      <div v-else-if="payment.status === 'Failed'" class="demo-actions">
        <button class="button button-primary" @click="retryPayment">Try again</button>
        <button class="button button-secondary" @click="navigateToHistory">Payment history</button>
      </div>

      <div v-else-if="payment.status === 'Cancelled'">
        <button class="button button-secondary" @click="navigateToInvoices">Back to invoices</button>
      </div>

      <div v-else>
        <button class="button button-primary" @click="navigateToHistory">View payment history</button>
      </div>

      <p class="panel-footnote">Payments are confirmed on the student's phone. Never share a PIN.</p>
    </section>
  </div>
</template>


<script setup>
import { ref, computed, onMounted } from 'vue'
import { useApi, formatRWF, statusBadge } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const payment = ref(null)

const isPaid = computed(() => payment.value?.status === 'Paid')
const isFailed = computed(() => payment.value?.status === 'Failed')

const title = computed(() => {
  if (!payment.value) return 'Payment'
  if (isPaid.value) return 'Payment confirmed'
  if (isFailed.value) return 'Payment not completed'
  return 'Approve payment on your phone'
})

const description = computed(() => {
  if (!payment.value) return ''
  if (isPaid.value) {
    return 'MoPay confirmed your payment. Your invoice balance and receipt have updated automatically.'
  }
  if (isFailed.value) {
    return 'MoPay could not complete the payment. Check your mobile-money balance before trying again.'
  }
  return 'Approve the MoPay request on your phone to complete your payment.'
})

const statusPanelClass = computed(() => {
  if (!payment.value) return ''
  if (isPaid.value) return 'status-panel-success'
  if (isFailed.value) return 'status-panel-error'
  return 'status-panel-pending'
})

const statusOrb = computed(() => {
  if (!payment.value) return '?'
  if (isPaid.value) return '✓'
  if (isFailed.value) return '!'
  return '↻'
})

const statusMessage = computed(() => {
  if (!payment.value) return ''
  if (isPaid.value) return 'The receipt is ready in your payment history.'
  if (isFailed.value) return 'No receipt is created until the payment is Paid.'
  return 'Waiting for MoPay confirmation…'
})

const simulatePaid = async () => {
  // In a real app, this would be handled by webhook
  // For demo, we'll just refresh the payment status
  try {
    const updated = await api.getPaymentById(payment.value.id)
    payment.value = updated
    emit('navigate', 'Payment History')
  } catch (error) {
    console.error('Failed to simulate payment:', error)
  }
}

const simulateFailed = async () => {
  try {
    await api.cancelPayment(payment.value.id)
    payment.value.status = 'Failed'
  } catch (error) {
    console.error('Failed to cancel payment:', error)
  }
}

const cancelPayment = async () => {
  try {
    await api.cancelPayment(payment.value.id)
    payment.value.status = 'Cancelled'
    emit('navigate', 'My Invoices')
  } catch (error) {
    console.error('Failed to cancel payment:', error)
  }
}

const retryPayment = () => {
  emit('navigate', 'Payment')
  localStorage.setItem('selectedInvoice', JSON.stringify({ id: payment.value.invoice_id }))
}

const navigateToHistory = () => {
  emit('navigate', 'Payment History')
}

const navigateToInvoices = () => {
  emit('navigate', 'My Invoices')
}

onMounted(() => {
  const currentPayment = localStorage.getItem('currentPayment')
  if (currentPayment) {
    payment.value = JSON.parse(currentPayment)
  }
})
</script>
