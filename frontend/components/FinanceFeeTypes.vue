<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">FINANCE PORTAL</p>
        <h1>Fee types</h1>
        <p class="dashboard-subtitle">Manage the fee catalogue used when issuing student invoices.</p>
      </div>
    </header>

    <section class="panel-card form-panel">
      <div class="section-heading">
        <div>
          <h2>Add a fee type</h2>
          <p>Define a standard charge for invoice creation.</p>
        </div>
      </div>
      <form @submit.prevent="handleCreateFeeType" class="form-grid">
        <label>
          Fee name
          <input v-model="createForm.name" placeholder="e.g. Laboratory fees" required>
        </label>
        <label>
          Programme
          <select v-model="createForm.programme_id" required>
            <option value="">Select programme</option>
            <option v-for="programme in programmes" :key="programme.id" :value="programme.id">
              {{ programme.name }}
            </option>
          </select>
        </label>
        <label>
          Semester
          <select v-model="createForm.semester_id" required>
            <option value="">Select semester</option>
            <option v-for="semester in semesters" :key="semester.id" :value="semester.id">
              {{ semester.name }}
            </option>
          </select>
        </label>
        <label>
          Default amount (RWF)
          <input v-model.number="createForm.amount" type="number" min="1" required>
        </label>
        <label>
          Effective from
          <input v-model="createForm.effective_from" type="date" required>
        </label>
        <div class="form-actions">
          <button class="button button-primary" type="submit">Add fee type</button>
        </div>
      </form>
    </section>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>Fee catalogue</h2>
          <p>{{ feeTypes.length }} fee types configured</p>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>FEE TYPE</th>
              <th>PROGRAMME</th>
              <th>SEMESTER</th>
              <th>DEFAULT AMOUNT</th>
              <th>EFFECTIVE FROM</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="feeTypes.length === 0">
              <td colspan="5" class="empty-state">No fee types found.</td>
            </tr>
            <tr v-for="fee in feeTypes" :key="fee.id">
              <td>
                <strong>{{ fee.name }}</strong>
                <small>{{ fee.id }}</small>
              </td>
              <td>{{ fee.programme_name || '—' }}</td>
              <td>{{ fee.semester_name || '—' }}</td>
              <td>{{ formatRWF(fee.amount) }}</td>
              <td>{{ formatDate(fee.effective_from) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="panel-footnote">Changing a fee type does not modify invoices already issued.</p>
    </section>
  </div>
</template>


<script setup>
import { ref, onMounted } from 'vue'
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
const feeTypes = ref([])
const programmes = ref([])
const semesters = ref([])
const loading = ref(true)

const createForm = ref({
  name: '',
  programme_id: '',
  semester_id: '',
  amount: 0,
  effective_from: ''
})

const handleCreateFeeType = async () => {
  try {
    const feeTypeData = {
      name: createForm.value.name,
      programme_id: createForm.value.programme_id,
      semester_id: createForm.value.semester_id,
      amount: createForm.value.amount,
      effective_from: createForm.value.effective_from
    }

    await api.createFeeType(feeTypeData)

    // Reset form
    createForm.value = {
      name: '',
      programme_id: '',
      semester_id: '',
      amount: 0,
      effective_from: ''
    }

    // Reload fee types
    await loadFeeTypes()
  } catch (error) {
    console.error('Failed to create fee type:', error)
    alert('Failed to create fee type: ' + error.message)
  }
}

const loadFeeTypes = async () => {
  try {
    feeTypes.value = await api.getFeeTypes()
  } catch (error) {
    console.error('Failed to load fee types:', error)
  }
}

const loadProgrammesAndSemesters = async () => {
  try {
    programmes.value = await api.getProgrammes()
    semesters.value = await api.getSemesters()
  } catch (error) {
    console.error('Failed to load programmes and semesters:', error)
  }
}

onMounted(() => {
  loadFeeTypes()
  loadProgrammesAndSemesters()
})
</script>
