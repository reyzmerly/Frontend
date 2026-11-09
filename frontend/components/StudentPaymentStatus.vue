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
        <strong>{{ payment.mopay_reference || payment.provider_ref || '—' }}</strong>
      </div>

      <div v-if="payment.status === 'Pending'" class="demo-actions">
        <p v-if="isLoading">Checking payment status…</p>
        <p v-else>Waiting for MoPay confirmation on your phone…</p>
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
import { ref, computed, onMounted, onUnmounted } from 'vue'
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
const pollingInterval = ref(null)
const isLoading = ref(false)

const isPaid = computed(() => payment.value?.status === 'Paid')
const isFailed = computed(() => payment.value?.status === 'Failed')
const isPending = computed(() => payment.value?.status === 'Pending')

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

const checkPaymentStatus = async () => {
  if (!payment.value?.id) {
    console.log('No payment ID to check')
    return
  }
  
  try {
    isLoading.value = true
    console.log('Checking status for payment ID:', payment.value.id)
    const updated = await api.getPaymentById(payment.value.id)
    console.log('Payment status updated:', updated)
    console.log('Current status:', updated?.status)
    console.log('Is Paid?', updated?.status === 'Paid')
    
    if (updated) {
      payment.value = updated
      
      // If payment is confirmed, stop polling and navigate
      if (updated.status === 'Paid') {
        console.log('Payment is Paid - stopping polling and navigating')
        stopPolling()
        localStorage.removeItem('currentPayment')
        setTimeout(() => {
          emit('navigate', 'Payment History')
        }, 2000)
      } else if (updated.status === 'Failed' || updated.status === 'Cancelled') {
        console.log('Payment is Failed or Cancelled - stopping polling')
        stopPolling()
      } else {
        console.log('Payment status:', updated.status, '- continuing to poll')
      }
    }
  } catch (error) {
    console.error('Failed to check payment status:', error)
  } finally {
    isLoading.value = false
  }
}

const startPolling = () => {
  if (pollingInterval.value) return
  
  // Check immediately
  checkPaymentStatus()
  
  // Then poll every 5 seconds
  pollingInterval.value = setInterval(() => {
    checkPaymentStatus()
  }, 5000)
}

const stopPolling = () => {
  if (pollingInterval.value) {
    clearInterval(pollingInterval.value)
    pollingInterval.value = null
  }
}

const cancelPayment = async () => {
  try {
    stopPolling()
    await api.cancelPayment(payment.value.id)
    payment.value.status = 'Cancelled'
    localStorage.removeItem('currentPayment')
    emit('navigate', 'My Invoices')
  } catch (error) {
    console.error('Failed to cancel payment:', error)
    alert('Failed to cancel payment: ' + error.message)
  }
}

const retryPayment = () => {
  stopPolling()
  localStorage.removeItem('currentPayment')
  emit('navigate', 'Payment')
  localStorage.setItem('selectedInvoice', JSON.stringify({ id: payment.value.invoice_id }))
}

const navigateToHistory = () => {
  stopPolling()
  emit('navigate', 'Payment History')
}

const navigateToInvoices = () => {
  stopPolling()
  emit('navigate', 'My Invoices')
}

onMounted(() => {
  const currentPayment = localStorage.getItem('currentPayment')
  if (currentPayment) {
    try {
      payment.value = JSON.parse(currentPayment)
      
      // If payment has an ID, start polling for status updates
      if (payment.value?.id && isPending.value) {
        startPolling()
      }
    } catch (error) {
      console.error('Failed to parse payment data:', error)
    }
  }
})

// Clean up polling when component unmounts
onUnmounted(() => {
  stopPolling()
})
</script>
