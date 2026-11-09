<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">FINANCE PORTAL</p>
        <h1>Manage invoices</h1>
        <p class="dashboard-subtitle">Create and track student invoices. Payments can only be confirmed by MoPay.</p>
      </div>
      <div class="page-heading-action">
        <button class="button button-primary" @click="showCreateForm = true">＋ Create invoice</button>
      </div>
    </header>

    <section v-if="showCreateForm" class="panel-card form-panel">
      <div class="section-heading">
        <div>
          <h2>Issue a new invoice</h2>
          <p>New invoices appear in the student portal immediately.</p>
        </div>
        <button class="icon-button" @click="showCreateForm = false" aria-label="Close">×</button>
      </div>
      <form @submit.prevent="handleCreateInvoice" class="form-grid">
        <label>
          Student
          <select v-model="createForm.student_id" required>
            <option value="">Select student</option>
            <option v-for="student in students" :key="student.id" :value="student.id">
              {{ student.full_name }} · {{ student.student_number }}
            </option>
          </select>
        </label>
        <label>
          Fee type
          <select v-model="createForm.fee_type_id" required>
            <option value="">Select fee type</option>
            <option v-for="fee in activeFeeTypes" :key="fee.id" :value="fee.id">
              {{ fee.name }} · {{ formatRWF(fee.amount) }}
            </option>
          </select>
        </label>
        <label>
          Invoice description
          <input v-model="createForm.description" placeholder="e.g. Tuition Fees — Semester II" required>
        </label>
        <label>
          Amount (RWF)
          <input v-model.number="createForm.amount" type="number" min="1" required>
        </label>
        <label>
          Due date
          <input v-model="createForm.due_date" type="date" required>
        </label>
        <div class="form-actions">
          <button class="button button-primary" type="submit">Issue invoice</button>
          <button class="button button-secondary" type="button" @click="showCreateForm = false">Cancel</button>
        </div>
      </form>
    </section>

    <section class="panel-card">
      <div class="toolbar">
        <div>
          <h2>All invoices</h2>
          <p>{{ filteredInvoices.length }} records · Amounts in RWF</p>
        </div>
        <div class="toolbar-controls">
          <input
            v-model="searchTerm"
            type="search"
            placeholder="Search student or invoice…"
            aria-label="Search invoices"
          >
          <select v-model="statusFilter" aria-label="Filter by status">
            <option value="">All statuses</option>
            <option>Paid</option>
            <option>Partially Paid</option>
            <option>Draft</option>
            <option>Issued</option>
            <option>Overdue</option>
          </select>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>INVOICE</th>
              <th>STUDENT</th>
              <th>DESCRIPTION</th>
              <th>AMOUNT</th>
              <th>BALANCE</th>
              <th>DUE DATE</th>
              <th>STATUS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="filteredInvoices.length === 0">
              <td colspan="7" class="empty-state">No invoices match your search.</td>
            </tr>
            <tr v-for="invoice in filteredInvoices" :key="invoice.id">
              <td>
                <strong>{{ invoice.invoice_number || invoice.id }}</strong>
                <small>Issued {{ formatDate(invoice.created_at) }}</small>
              </td>
              <td>
                <strong>{{ invoice.student?.full_name || invoice.student_name || '—' }}</strong>
                <small>{{ invoice.student?.student_number || invoice.student_id || '—' }}</small>
              </td>
              <td>{{ invoice.description }}</td>
              <td>{{ formatRWF(invoice.total_amount) }}</td>
              <td><strong>{{ formatRWF(invoice.balance || (invoice.total_amount - invoice.paid_amount)) }}</strong></td>
              <td>{{ formatDate(invoice.due_date) }}</td>
              <td v-html="statusBadge(invoice.status)"></td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="panel-footnote">Drafts and issued totals are separate. A payment remains read-only until MoPay confirms it.</p>
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
const students = ref([])
const feeTypes = ref([])
const loading = ref(true)

const showCreateForm = ref(false)
const searchTerm = ref('')
const statusFilter = ref('')

const createForm = ref({
  student_id: '',
  fee_type_id: '',
  description: '',
  amount: 0,
  due_date: ''
})

const activeFeeTypes = computed(() => {
  return feeTypes.value.filter(fee => fee.active !== false)
})

const filteredInvoices = computed(() => {
  const term = searchTerm.value.toLowerCase()
  const status = statusFilter.value

  return invoices.value.filter(invoice => {
    const studentName = invoice.student?.full_name || invoice.student_name || ''
    const studentNumber = invoice.student?.student_number || invoice.student_id || ''
    const invoiceNum = invoice.invoice_number || invoice.id || ''
    
    const matchesSearch = `${invoiceNum} ${studentName} ${studentNumber} ${invoice.description}`.toLowerCase().includes(term)
    const matchesStatus = !status || invoice.status === status
    return matchesSearch && matchesStatus
  })
})

const handleCreateInvoice = async () => {
  try {
    const invoiceData = {
      student_id: createForm.value.student_id,
      fee_type_id: createForm.value.fee_type_id,
      description: createForm.value.description,
      total_amount: createForm.value.amount,
      due_date: createForm.value.due_date
    }

    await api.createInvoice(invoiceData)
    showCreateForm.value = false
    // Reset form
    createForm.value = {
      student_id: '',
      fee_type_id: '',
      description: '',
      amount: 0,
      due_date: ''
    }
    // Reload invoices
    await loadInvoices()
  } catch (error) {
    console.error('Failed to create invoice:', error)
    alert('Failed to create invoice: ' + error.message)
  }
}

const loadInvoices = async () => {
  try {
    invoices.value = await api.getInvoices()
  } catch (error) {
    console.error('Failed to load invoices:', error)
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const [invoicesData, studentsData, feeTypesData] = await Promise.all([
      api.getInvoices(),
      api.getStudents(),
      api.getFeeTypes()
    ])
    invoices.value = invoicesData || []
    students.value = studentsData || []
    feeTypes.value = feeTypesData || []
  } catch (error) {
    console.error('Failed to load finance invoices data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
