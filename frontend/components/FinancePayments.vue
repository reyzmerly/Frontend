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
                <strong>{{ payment.student?.full_name || payment.student_name || '—' }}</strong>
                <small>{{ payment.student?.student_number || payment.student_id || '—' }}</small>
              </td>
              <td>{{ payment.invoice_number || payment.invoice_id }}</td>
              <td>{{ payment.mopay_reference || payment.provider_ref || '—' }}</td>
              <td>{{ formatRWF(payment.amount) }}</td>
              <td v-html="statusBadge(payment.status)"></td>
              <td>
                <button
                  class="button button-quiet button-small"
                  @click="addNote(payment)"
                >
                  Add note
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <!-- Note Modal -->
    <div v-if="showNoteModal" class="modal-overlay" @click.self="showNoteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Add Payment Note</h2>
          <button class="icon-button" @click="showNoteModal = false">×</button>
        </div>
        <div class="modal-body">
          <p class="note-info">
            <strong>Payment:</strong> {{ selectedPayment?.invoice_number || selectedPayment?.invoice_id }}
            <br>
            <strong>Amount:</strong> {{ formatRWF(selectedPayment?.amount) }}
            <br>
            <strong>Reference:</strong> {{ selectedPayment?.mopay_reference || selectedPayment?.provider_ref || '—' }}
          </p>
          <label>
            Note
            <textarea
              v-model="noteText"
              placeholder="Enter your follow-up note..."
              rows="4"
              maxlength="500"
            ></textarea>
            <small>{{ noteText.length }}/500 characters</small>
          </label>
          <div class="modal-actions">
            <button class="button button-secondary" @click="showNoteModal = false">Cancel</button>
            <button class="button button-primary" @click="submitNote" :disabled="!noteText.trim()">Save Note</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Success Modal -->
    <div v-if="showSuccessModal" class="modal-overlay" @click.self="showSuccessModal = false">
      <div class="modal-content modal-success">
        <div class="modal-body">
          <div class="success-icon">✓</div>
          <h2>Note Added Successfully</h2>
          <p>Your follow-up note has been saved to the payment record.</p>
          <div class="modal-actions">
            <button class="button button-primary" @click="showSuccessModal = false">Close</button>
          </div>
        </div>
      </div>
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
const payments = ref([])
const loading = ref(true)

const searchTerm = ref('')
const statusFilter = ref('')

const showNoteModal = ref(false)
const selectedPayment = ref(null)
const noteText = ref('')
const showSuccessModal = ref(false)

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
    const studentName = payment.student?.full_name || payment.student_name || ''
    const studentNumber = payment.student?.student_number || payment.student_id || ''
    const invoiceNum = payment.invoice_number || payment.invoice_id || ''
    const ref = payment.mopay_reference || payment.provider_ref || ''
    
    const matchesSearch = `${studentName} ${studentNumber} ${invoiceNum} ${ref}`.toLowerCase().includes(search)
    const matchesStatus = !status || payment.status === status
    return matchesSearch && matchesStatus
  })
})

const formatTime = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const addNote = (payment) => {
  selectedPayment.value = payment
  noteText.value = ''
  showNoteModal.value = true
}

const submitNote = async () => {
  if (!noteText.value.trim()) return

  try {
    await api.createPaymentNote(selectedPayment.value.id, noteText.value.trim())
    showNoteModal.value = false
    showSuccessModal.value = true
    await loadPayments()
  } catch (error) {
    console.error('Failed to add note:', error)
    alert('Failed to add note: ' + error.message)
  }
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

<style scoped>
.dashboard-subtitle {
  font-size: 15px;
  color: #6b7280;
  margin: 0 0 20px 0;
}

.note-info {
  background: #f9fafb;
  padding: 15px;
  border-radius: 6px;
  margin-bottom: 20px;
  font-size: 14px;
  line-height: 1.6;
}

.note-info strong {
  color: #374151;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.modal-content {
  background: white;
  border-radius: 8px;
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h2 {
  margin: 0;
  font-size: 20px;
}

.modal-body {
  padding: 20px;
}

.modal-body label {
  display: block;
  margin-bottom: 20px;
}

.modal-body textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-family: inherit;
  font-size: 14px;
  resize: vertical;
  margin-top: 5px;
}

.modal-body small {
  display: block;
  margin-top: 5px;
  color: #6b7280;
  font-size: 12px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.icon-button {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}

.icon-button:hover {
  background: #f3f4f6;
}

.modal-success {
  text-align: center;
  max-width: 400px;
}

.modal-success .modal-body {
  padding: 40px 20px;
}

.success-icon {
  width: 60px;
  height: 60px;
  background: #10b981;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: bold;
  margin: 0 auto 20px auto;
}

.modal-success h2 {
  margin: 0 0 10px 0;
  font-size: 20px;
  color: #374151;
}

.modal-success p {
  margin: 0 0 20px 0;
  color: #6b7280;
  font-size: 14px;
}

.modal-success .modal-actions {
  justify-content: center;
}
</style>
