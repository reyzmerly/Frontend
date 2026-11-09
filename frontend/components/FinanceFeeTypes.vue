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
          Category
          <select v-model="createForm.category" required>
            <option>Academic</option>
            <option>Administrative</option>
            <option>Facilities</option>
            <option>Campus</option>
          </select>
        </label>
        <label>
          Default amount (RWF)
          <input v-model.number="createForm.amount" type="number" min="1" required>
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
              <th>CATEGORY</th>
              <th>DEFAULT AMOUNT</th>
              <th>STATUS</th>
              <th>ACTION</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="fee in feeTypes" :key="fee.id">
              <td>
                <strong>{{ fee.name }}</strong>
                <small>{{ fee.id }}</small>
              </td>
              <td>{{ fee.category }}</td>
              <td>{{ formatRWF(fee.amount) }}</td>
              <td>
                <span :class="['status-pill', fee.active ? 'status-paid' : 'status-overdue']">
                  {{ fee.active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td>
                <button
                  class="button button-quiet button-small"
                  @click="toggleFeeType(fee)"
                >
                  {{ fee.active ? 'Deactivate' : 'Activate' }}
                </button>
              </td>
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
import { useApi, formatRWF } from '~/composables/useApi'

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
const loading = ref(true)

const createForm = ref({
  name: '',
  category: 'Academic',
  amount: 0
})

const handleCreateFeeType = async () => {
  try {
    const feeTypeData = {
      name: createForm.value.name,
      category: createForm.value.category,
      amount: createForm.value.amount,
      active: true
    }

    await api.createFeeType(feeTypeData)

    // Reset form
    createForm.value = {
      name: '',
      category: 'Academic',
      amount: 0
    }

    // Reload fee types
    await loadFeeTypes()
  } catch (error) {
    console.error('Failed to create fee type:', error)
    alert('Failed to create fee type: ' + error.message)
  }
}

const toggleFeeType = async (fee) => {
  try {
    const updatedData = {
      ...fee,
      active: !fee.active
    }

    await api.updateFeeType(fee.id, updatedData)

    // Reload fee types
    await loadFeeTypes()
  } catch (error) {
    console.error('Failed to toggle fee type:', error)
    alert('Failed to update fee type: ' + error.message)
  }
}

const loadFeeTypes = async () => {
  try {
    feeTypes.value = await api.getFeeTypes()
  } catch (error) {
    console.error('Failed to load fee types:', error)
  }
}

onMounted(() => {
  loadFeeTypes()
})
</script>
