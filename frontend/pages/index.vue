<template>
  <div>
    <!-- Check if user is logged in -->
    <div v-if="!user" class="login-container">
      <section class="login-brand">
        <img class="university-logo" src="/public/assets/images/UoK.png" alt="University of Kigali logo">
        <p class="eyebrow">UNIVERSITY OF KIGALI</p>
        <h1>Your finances.<br>Clearly in view.</h1>
        <br>
        <h3>Invoices and Payments Records<br>Management System</h3>
        <p class="brand-description">
          View invoices. Pay securely with MoPay.<br>
          Keep every receipt in one place.
        </p>
        <div class="login-benefits">
          <p>01 View your invoices</p>
          <p>02 Make secure payments</p>
          <p>03 Access your receipts</p>
        </div>
        <footer class="login-footer">TEAM 3 UNSTOPPABLE</footer>
      </section>

      <section class="login-form-section">
        <form @submit.prevent="handleLogin" class="login-form">
          <h2>Welcome Back</h2>
          <p class="eyebrow">Sign in with your university account.</p>
          <p class="form-description">
            Enter your university credentials to continue.
          </p>
          <label for="identity">Email or Student ID</label>
          <input
            type="text"
            id="identity"
            v-model="form.identifier"
            placeholder="Enter your email or student ID"
            required
          >
          <label for="password">Password</label>
          <input
            type="password"
            id="password"
            v-model="form.password"
            placeholder="Enter your password"
            required
          >
          <button type="submit" class="login-button">
            Sign in
            <span>→</span>
          </button>
          <p v-if="error" class="error-message">{{ error }}</p>
          <p class="demo-notice">
            Your role determines your workspace. Students see only their own records; staff access is authorized.
          </p>
          <button type="button" class="forgot-password">Forgot password?</button>
          <p class="demo-credentials">Demo accounts: student, finance or admin · password: 123</p>
        </form>
      </section>
    </div>

    <!-- Application Workspace -->
    <div v-else class="application">
      <aside class="sidebar">
        <div class="sidebar-brand">
          <img class="university-logo small-logo" src="/public/assets/images/UoK.png" alt="University of Kigali logo">
          <div>
            <strong>University of Kigali</strong>
            <small>PAYMENT SYSTEM</small>
          </div>
        </div>
        <p class="sidebar-label">{{ portalLabel }}</p>
        <nav class="navigation-menu">
          <button
            v-for="page in navigationPages"
            :key="page"
            :class="['nav-link', currentPage === page ? 'active' : '']"
            @click="currentPage = page"
          >
            {{ page }}
          </button>
        </nav>
        <button @click="handleLogout" class="logout-button">
          Sign out
        </button>
      </aside>

      <section class="workspace">
        <header class="topbar">
          <div>
            <p id="page-breadcrumb">{{ currentPage }}</p>
            <small class="term-label">University of Kigali · Semester 1, 2026/27</small>
          </div>
          <div class="user-profile">
            <div class="user-avatar">{{ initials }}</div>
            <div>
              <strong>{{ user.full_name }}</strong>
              <small>{{ user.role }}</small>
            </div>
          </div>
        </header>
        <div class="page-content">
          <!-- Dynamic content based on role and page -->
          <div v-if="currentPage === 'Dashboard'">
            <div v-if="user.role === 'student'">
              <ClientOnly>
                <StudentDashboard :user="user" :current-page="currentPage" @navigate="handleNavigate" />
              </ClientOnly>
            </div>
            <div v-else-if="user.role === 'finance'">
              <ClientOnly>
                <FinanceDashboard :user="user" :current-page="currentPage" @navigate="handleNavigate" />
              </ClientOnly>
            </div>
            <div v-else-if="user.role === 'admin'">
              <ClientOnly>
                <AdminDashboard :user="user" :current-page="currentPage" @navigate="handleNavigate" />
              </ClientOnly>
            </div>
          </div>
          <div v-else-if="currentPage === 'My Invoices' && user.role === 'student'">
            <ClientOnly>
              <StudentInvoices :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Payment History' && user.role === 'student'">
            <ClientOnly>
              <StudentHistory :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Payment' && user.role === 'student'">
            <ClientOnly>
              <StudentPayment :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Payment Status' && user.role === 'student'">
            <ClientOnly>
              <StudentPaymentStatus :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Manage Invoices' && user.role === 'finance'">
            <ClientOnly>
              <FinanceInvoices :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Fee Types' && user.role === 'finance'">
            <ClientOnly>
              <FinanceFeeTypes :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Payments monitor' && user.role === 'finance'">
            <ClientOnly>
              <FinancePayments :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Financial Reports' && user.role === 'finance'">
            <ClientOnly>
              <FinanceReports :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Programmes' && user.role === 'finance'">
            <ClientOnly>
              <FinanceProgrammes :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Semesters' && user.role === 'finance'">
            <ClientOnly>
              <FinanceSemesters :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Roles and access' && user.role === 'admin'">
            <ClientOnly>
              <AdminUsers :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
          <div v-else-if="currentPage === 'Audit Log' && user.role === 'admin'">
            <ClientOnly>
              <AdminAudit :user="user" :current-page="currentPage" @navigate="handleNavigate" />
            </ClientOnly>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useHead } from '#imports'
import StudentDashboard from '~/components/StudentDashboard.vue'
import StudentInvoices from '~/components/StudentInvoices.vue'
import StudentHistory from '~/components/StudentHistory.vue'
import StudentPayment from '~/components/StudentPayment.vue'
import StudentPaymentStatus from '~/components/StudentPaymentStatus.vue'
import FinanceDashboard from '~/components/FinanceDashboard.vue'
import FinanceInvoices from '~/components/FinanceInvoices.vue'
import FinanceFeeTypes from '~/components/FinanceFeeTypes.vue'
import FinancePayments from '~/components/FinancePayments.vue'
import FinanceReports from '~/components/FinanceReports.vue'
import FinanceProgrammes from '~/components/FinanceProgrammes.vue'
import FinanceSemesters from '~/components/FinanceSemesters.vue'
import AdminUsers from '~/components/AdminUsers.vue'
import AdminAudit from '~/components/AdminAudit.vue'

useHead({
  title: 'University of Kigali - Invoice & Payment Management'
})

const router = useRouter()

const form = ref({
  identifier: '',
  password: ''
})
const error = ref('')
const user = ref(null)
const currentPage = ref('Dashboard')

const portalLabel = computed(() => {
  if (!user.value) return ''
  const role = user.value.role
  return role === 'admin' ? 'ADMIN PORTAL' : role.toUpperCase() + ' PORTAL'
})

const initials = computed(() => {
  if (!user.value) return ''
  return user.value.full_name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
})

const navigationPages = computed(() => {
  if (!user.value) return []
  const role = user.value.role
  if (role === 'student') {
    return ['Dashboard', 'My Invoices', 'Payment History']
  } else if (role === 'finance') {
    return ['Dashboard', 'Manage Invoices', 'Fee Types', 'Payments monitor', 'Financial Reports', 'Programmes', 'Semesters']
  } else if (role === 'admin') {
    return ['Dashboard', 'Roles and access', 'Audit Log']
  }
  return []
})

const handleNavigate = (page) => {
  currentPage.value = page
}

const handleLogin = async () => {
  error.value = ''
  try {
    const response = await fetch('http://localhost:9107/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        identifier: form.value.identifier,
        password: form.value.password
      })
    })

    const data = await response.json()

    // Store token and user info
    localStorage.setItem('token', data.access_token)
    localStorage.setItem('user', JSON.stringify(data.user))
    user.value = data.user
  } catch (e) {
    error.value = 'Invalid credentials'
  }
}

const handleLogout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  user.value = null
  currentPage.value = 'Dashboard'
}

onMounted(() => {
  const storedUser = localStorage.getItem('user')
  if (storedUser) {
    user.value = JSON.parse(storedUser)
  }
})
</script>

<style>
/* Use the existing CSS files */
@import url('~/assets/css/main.css');
</style>
