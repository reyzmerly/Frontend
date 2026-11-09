<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">ADMIN PORTAL</p>
        <h1>Roles and access</h1>
        <p class="dashboard-subtitle">Review role assignments and access boundaries. Payment statuses remain controlled by MoPay.</p>
      </div>
      <button class="button button-secondary" @click="navigateTo('Admin overview')">
        ← Back to Dashboard
      </button>
    </header>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>User roles</h2>
          <p>{{ users.length }} accounts · Access should follow least-privilege principles</p>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table">
          <thead>
            <tr>
              <th>USER</th>
              <th>LOGIN ID</th>
              <th>ROLE</th>
              <th>ACCESS STATUS</th>
              <th>CREATED</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="account in users" :key="account.id">
              <td>
                <strong>{{ account.full_name }}</strong>
                <small>{{ account.id }}</small>
              </td>
              <td>{{ account.email }}</td>
              <td>
                <select
                  v-model="account.role"
                  class="inline-select"
                  @change="updateUserAccess(account)"
                  :aria-label="`Role for ${account.full_name}`"
                >
                  <option value="student">Student</option>
                  <option value="finance">Finance</option>
                  <option value="admin">Administrator</option>
                </select>
              </td>
              <td>
                <select
                  v-model="account.access_status"
                  class="inline-select"
                  @change="updateUserAccess(account)"
                  :aria-label="`Access status for ${account.full_name}`"
                >
                  <option value="active">Active</option>
                  <option value="suspended">Suspended</option>
                  <option value="invited">Invited</option>
                </select>
              </td>
              <td>
                <small>{{ formatDate(account.created_at) }}</small>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>Role permissions</h2>
          <p>High-level access matrix</p>
        </div>
      </div>
      <div class="table-scroll">
        <table class="data-table permission-table">
          <thead>
            <tr>
              <th>CAPABILITY</th>
              <th>STUDENT</th>
              <th>FINANCE</th>
              <th>ADMIN</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(perm, index) in permissions" :key="index">
              <td>{{ perm[0] }}</td>
              <td v-html="permissionMark(perm[1])"></td>
              <td v-html="permissionMark(perm[2])"></td>
              <td v-html="permissionMark(perm[3])"></td>
            </tr>
          </tbody>
        </table>
      </div>
      <p class="panel-footnote">Admin controls role assignments and can view audit history. Admin does not have manual payment controls.</p>
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
const users = ref([])
const loading = ref(true)

const navigateTo = (page) => {
  emit('navigate', page)
}

const permissions = [
  ['Own invoices, MoPay checkout and Paid receipts', true, false, false],
  ['Create invoices and manage fee types', false, true, false],
  ['View payment monitor and financial reports', false, true, false],
  ['Manage roles and read-only audit log', false, false, true]
]

const permissionMark = (enabled) => {
  return enabled
    ? `<span class="permission-check" aria-label="Allowed">✓</span>`
    : `<span class="permission-none" aria-label="Not allowed">—</span>`
}

const updateUserAccess = async (account) => {
  try {
    await api.updateUser(account.id, {
      role: account.role,
      access_status: account.access_status
    })
    // Reload users
    await loadUsers()
  } catch (error) {
    console.error('Failed to update user access:', error)
    alert('Failed to update user access: ' + error.message)
  }
}

const loadUsers = async () => {
  try {
    users.value = await api.getUsers()
  } catch (error) {
    console.error('Failed to load users:', error)
  }
}

onMounted(() => {
  loadUsers()
})
</script>
