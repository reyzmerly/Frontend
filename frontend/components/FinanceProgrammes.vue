<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">FINANCE PORTAL</p>
        <h1>Academic programmes</h1>
        <p class="dashboard-subtitle">View and manage academic programmes across the university.</p>
      </div>
    </header>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>All programmes</h2>
          <p>{{ programmes.length }} programmes registered</p>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>PROGRAMME NAME</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="programmes.length === 0">
              <td colspan="2" class="empty-state">No programmes found.</td>
            </tr>
            <tr v-for="programme in programmes" :key="programme.id">
              <td>{{ programme.id }}</td>
              <td>
                <strong>{{ programme.name }}</strong>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="panel-footnote">Programmes are managed by academic administration. Finance staff can view but not modify programmes.</p>
    </section>
  </div>
</template>


<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const programmes = ref([])
const loading = ref(true)

const loadProgrammes = async () => {
  try {
    programmes.value = await api.getProgrammes()
  } catch (error) {
    console.error('Failed to load programmes:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadProgrammes()
})
</script>
