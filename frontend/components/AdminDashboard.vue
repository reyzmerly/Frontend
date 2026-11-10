<template>
  <div>
    <header class="page-heading">
      <div>
        <p class="dashboard-eyebrow">ADMIN PORTAL</p>
        <h1>Admin overview</h1>
        <p class="dashboard-subtitle">Manage system-wide access and view audit logs.</p>
      </div>
    </header>

    <section class="metric-grid metric-grid-3">
      <article class="metric-card">
        <span class="metric-icon icon-burgundy">▤</span>
        <p>Total users</p>
        <strong>{{ users.length }}</strong>
        <small>Registered accounts</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-green">✓</span>
        <p>Active accounts</p>
        <strong>{{ activeUsers }}</strong>
        <small>Currently active</small>
      </article>
      <article class="metric-card">
        <span class="metric-icon icon-gold">◷</span>
        <p>Audit events</p>
        <strong>{{ auditLog.length }}</strong>
        <small>Recorded activities</small>
      </article>
    </section>

    <section class="panel-card">
      <div class="section-heading">
        <div>
          <h2>Quick actions</h2>
          <p>Common administrative tasks</p>
        </div>
      </div>
      <div class="action-grid">
        <button class="button button-secondary" @click="navigateTo('Roles and access')">
          Manage user roles
        </button>
        <button class="button button-secondary" @click="navigateTo('Audit Log')">
          View audit log
        </button>
        <button class="button button-primary" @click="showInviteModal = true">
          Invite new user
        </button>
      </div>
    </section>

    <!-- Invite User Modal -->
    <div v-if="showInviteModal" class="modal-overlay" @click.self="showInviteModal = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Invite New User</h2>
          <button class="icon-button" @click="showInviteModal = false">×</button>
        </div>
        <div class="modal-body">
          <label>
            Full name
            <input v-model="inviteForm.full_name" type="text" placeholder="Enter full name" required>
          </label>
          <label>
            Email
            <input v-model="inviteForm.email" type="email" placeholder="Enter email address" required>
          </label>
          <label>
            Role
            <select v-model="inviteForm.role" required @change="handleRoleChange">
              <option value="">Select role</option>
              <option value="student">Student</option>
              <option value="finance">Finance</option>
              <option value="admin">Administrator</option>
            </select>
          </label>
          <label v-if="inviteForm.role === 'student'">
            Student number
            <input v-model="inviteForm.student_number" type="text" placeholder="e.g. UOK/2024/0001" required>
          </label>
          <label v-if="inviteForm.role === 'student'">
            Programme
            <select v-model="inviteForm.programme_id" required>
              <option value="">Select programme</option>
              <option v-for="programme in programmes" :key="programme.id" :value="programme.id">
                {{ programme.name }}
              </option>
            </select>
          </label>
          <label>
            Password
            <input v-model="inviteForm.password" type="password" placeholder="Enter password (min 8 characters)" required minlength="8">
          </label>
          <div class="modal-actions">
            <button class="button button-secondary" @click="showInviteModal = false">Cancel</button>
            <button class="button button-primary" @click="inviteUser" :disabled="!isInviteFormValid">Send Invite</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Success Modal -->
    <div v-if="showSuccessModal" class="modal-overlay" @click.self="showSuccessModal = false">
      <div class="modal-content modal-success">
        <div class="modal-body">
          <div class="success-icon">✓</div>
          <h2>User Invited Successfully</h2>
          <p>The new user has been created and can now log in with their credentials.</p>
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
import { useApi } from '~/composables/useApi'

const props = defineProps({
  user: Object,
  currentPage: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const api = useApi()
const users = ref([])
const auditLog = ref([])
const programmes = ref([])
const loading = ref(true)

const showInviteModal = ref(false)
const showSuccessModal = ref(false)
const inviteForm = ref({
  full_name: '',
  email: '',
  role: '',
  password: '',
  student_number: '',
  programme_id: ''
})

const isInviteFormValid = computed(() => {
  const baseValid = inviteForm.value.full_name.trim() &&
                    inviteForm.value.email.trim() &&
                    inviteForm.value.role &&
                    inviteForm.value.password.length >= 8

  if (!baseValid) return false

  // For students, require student_number and programme_id
  if (inviteForm.value.role === 'student') {
    return inviteForm.value.student_number.trim() && inviteForm.value.programme_id
  }

  return true
})

const activeUsers = computed(() => {
  return users.value.filter(user => user.access_status === 'active').length
})

const navigateTo = (page) => {
  emit('navigate', page)
}

const handleRoleChange = () => {
  // Clear student-specific fields when role changes
  if (inviteForm.value.role !== 'student') {
    inviteForm.value.student_number = ''
    inviteForm.value.programme_id = ''
  }
}

const inviteUser = async () => {
  try {
    const userData = {
      full_name: inviteForm.value.full_name,
      email: inviteForm.value.email,
      role: inviteForm.value.role,
      password: inviteForm.value.password
    }

    // Add student-specific fields if role is student
    if (inviteForm.value.role === 'student') {
      userData.student_number = inviteForm.value.student_number
      userData.programme_id = inviteForm.value.programme_id
    }

    await api.createUser(userData)
    showInviteModal.value = false
    showSuccessModal.value = true
    // Reset form
    inviteForm.value = {
      full_name: '',
      email: '',
      role: '',
      password: '',
      student_number: '',
      programme_id: ''
    }
    // Reload users to update metrics
    await loadData()
  } catch (error) {
    console.error('Failed to invite user:', error)
    alert('Failed to invite user: ' + error.message)
  }
}

const loadData = async () => {
  try {
    loading.value = true
    const [usersData, auditData, programmesData] = await Promise.all([
      api.getUsers(),
      api.getAuditLogs(),
      api.getProgrammes()
    ])
    users.value = usersData || []
    auditLog.value = auditData || []
    programmes.value = programmesData || []
  } catch (error) {
    console.error('Failed to load admin dashboard data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.action-grid {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
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
  margin-bottom: 15px;
}

.modal-body input,
.modal-body select {
  width: 100%;
  padding: 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  margin-top: 5px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 20px;
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
