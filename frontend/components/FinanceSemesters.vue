<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">FINANCE PORTAL</p>
        <h1>Academic semesters</h1>
        <p class="dashboard-subtitle">View and manage academic semesters for fee collection.</p>
      </div>
    </header>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>All semesters</h2>
          <p>{{ semesters.length }} semesters configured</p>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>SEMESTER NAME</th>
              <th>START DATE</th>
              <th>END DATE</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="semesters.length === 0">
              <td colspan="4" class="empty-state">No semesters found.</td>
            </tr>
            <tr v-for="semester in semesters" :key="semester.id">
              <td>{{ semester.id }}</td>
              <td>
                <strong>{{ semester.name }}</strong>
              </td>
              <td>{{ formatDate(semester.starts_on) }}</td>
              <td>{{ formatDate(semester.ends_on) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="panel-footnote">Semesters are managed by academic administration. Finance staff can view but not modify semesters.</p>
    </section>
  </div>
</template>


<script setup>
import { ref, onMounted } from 'vue'
import { useApi, formatDate } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const semesters = ref([])
const loading = ref(true)

const loadSemesters = async () => {
  try {
    semesters.value = await api.getSemesters()
  } catch (error) {
    console.error('Failed to load semesters:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadSemesters()
})
</script>
