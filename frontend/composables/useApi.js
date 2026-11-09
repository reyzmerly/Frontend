const API_BASE_URL = 'http://localhost:9107/api/v1'

// Get JWT token from localStorage (client-side only)
function getAuthToken() {
  if (typeof window === 'undefined') return null
  const token = localStorage.getItem('token')
  return token || null
}

// Generic API call helper
export function useApi() {
  const apiCall = async (endpoint, options = {}) => {
    const token = getAuthToken()
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers
    }

    if (token) {
      headers['Authorization'] = `Bearer ${token}`
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, {
      ...options,
      headers
    })

    if (!response.ok) {
      const errorText = await response.text()
      let error
      try {
        error = JSON.parse(errorText)
      } catch {
        error = { message: errorText || 'Request failed' }
      }
      throw new Error(error.message || error.code || 'Request failed')
    }

    // Handle response body - 202 Accepted might have empty body
    const text = await response.text()
    
    if (!text || text.trim() === '') {
      console.log('Empty response body (likely 202 Accepted)')
      return { status: response.status, empty: true }
    }
    
    try {
      return JSON.parse(text)
    } catch (e) {
      console.error('JSON parse error:', e)
      throw new Error('Invalid JSON response from server')
    }
  }

  // Authentication
  const login = async (identifier, password) => {
    const response = await apiCall('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ identifier, password })
    })

    // Store token and user info
    localStorage.setItem('token', response.access_token)
    localStorage.setItem('user', JSON.stringify(response.user))

    return response
  }

  const logout = async () => {
    try {
      await apiCall('/auth/logout', {
        method: 'POST'
      })
    } catch (error) {
      console.error('Logout error:', error)
    }
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // Invoices
  const getInvoices = async (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    const response = await apiCall(`/invoices${queryString ? `?${queryString}` : ''}`)
    return response.data || []
  }

  const getInvoiceById = async (invoiceId) => {
    const response = await apiCall(`/invoices/${invoiceId}`)
    // Single object endpoints return data directly, not wrapped in { data: ... }
    return response.data || response
  }

  const createInvoice = async (invoiceData) => {
    const response = await apiCall('/invoices', {
      method: 'POST',
      body: JSON.stringify(invoiceData)
    })
    return response.data
  }

  const issueInvoice = async (invoiceId) => {
    const response = await apiCall(`/invoices/${invoiceId}/issue`, {
      method: 'POST',
      body: JSON.stringify({ action: 'issue' })
    })
    return response.data || response
  }

  const updateInvoice = async (invoiceId, invoiceData) => {
    const response = await apiCall(`/invoices/${invoiceId}`, {
      method: 'PUT',
      body: JSON.stringify(invoiceData)
    })
    return response.data
  }

  // Payments
  const getPayments = async (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    const response = await apiCall(`/payments${queryString ? `?${queryString}` : ''}`)
    return response.data || []
  }

  const getPaymentById = async (paymentId) => {
    const response = await apiCall(`/payments/${paymentId}`)
    // Single object endpoints return data directly, not wrapped in { data: ... }
    return response.data || response
  }

  const createPayment = async (paymentData) => {
    const response = await apiCall(`/invoices/${paymentData.invoice_id}/payments`, {
      method: 'POST',
      body: JSON.stringify({
        amount: paymentData.amount,
        mobile_number: paymentData.phone
      })
    })
    // Handle 202 Accepted response with payment data
    if (!response || Object.keys(response).length === 0) {
      return { success: true }
    }
    return response.data || response
  }

  const cancelPayment = async (paymentId) => {
    await apiCall(`/payments/${paymentId}/cancel`, {
      method: 'POST'
    })
  }

  const createPaymentNote = async (paymentId, note) => {
    const response = await apiCall(`/payments/${paymentId}/notes`, {
      method: 'POST',
      body: JSON.stringify({ note })
    })
    return response.data || response
  }

  // Fee Types
  const getFeeTypes = async () => {
    const response = await apiCall('/fee-types')
    // Backend returns direct array, not wrapped in { data: ... }
    return Array.isArray(response) ? response : (response.data || [])
  }

  const createFeeType = async (feeTypeData) => {
    const response = await apiCall('/fee-types', {
      method: 'POST',
      body: JSON.stringify(feeTypeData)
    })
    return response.data
  }

  const updateFeeType = async (feeTypeId, feeTypeData) => {
    const response = await apiCall(`/fee-types/${feeTypeId}`, {
      method: 'PUT',
      body: JSON.stringify(feeTypeData)
    })
    return response.data
  }

  // Students
  const getStudents = async (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    const response = await apiCall(`/finance/students${queryString ? `?${queryString}` : ''}`)
    // Backend returns direct array, not wrapped in { data: ... }
    return Array.isArray(response) ? response : (response.data || [])
  }

  // Finance Reports
  const getAccountsReport = async () => {
    const response = await apiCall('/finance/reports/accounts')
    return response.data
  }

  const getReportSummary = async () => {
    const response = await apiCall('/finance/reports/summary')
    return response.data
  }

  const exportFinanceReport = async (format = 'pdf') => {
    const token = getAuthToken()
    const response = await fetch(`${API_BASE_URL}/finance/reports/export?format=${format}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) {
      throw new Error('Failed to export report')
    }

    return response.blob()
  }

  // Finance Dashboard
  const getFinanceDashboard = async () => {
    const response = await apiCall('/finance/dashboard')
    return response.data
  }

  // Finance Programmes
  const getProgrammes = async () => {
    const response = await apiCall('/finance/programmes')
    // Backend returns direct array, not wrapped in { data: ... }
    return Array.isArray(response) ? response : (response.data || [])
  }

  // Finance Semesters
  const getSemesters = async () => {
    const response = await apiCall('/finance/semesters')
    // Backend returns direct array, not wrapped in { data: ... }
    return Array.isArray(response) ? response : (response.data || [])
  }

  // Users (Admin)
  const getUsers = async () => {
    const response = await apiCall('/admin/users')
    // Backend returns direct array, not wrapped in { data: ... }
    return Array.isArray(response) ? response : (response.data || [])
  }

  const createUser = async (userData) => {
    const response = await apiCall('/admin/users', {
      method: 'POST',
      body: JSON.stringify(userData)
    })
    return response.data
  }

  const getUserById = async (userId) => {
    const response = await apiCall(`/admin/users/${userId}`)
    return response.data
  }

  const updateUser = async (userId, userData) => {
    const response = await apiCall(`/admin/users/${userId}/access`, {
      method: 'PATCH',
      body: JSON.stringify(userData)
    })
    return response.data
  }

  // Audit Logs (Admin)
  const getAuditLogs = async (params = {}) => {
    const queryString = new URLSearchParams(params).toString()
    const response = await apiCall(`/admin/audit-logs${queryString ? `?${queryString}` : ''}`)
    // Backend returns { data, page, page_size, total }
    return response.data || []
  }

  // Receipts
  const getReceipt = async (receiptId) => {
    const response = await apiCall(`/receipts/${receiptId}`)
    // Single object endpoints return data directly, not wrapped in { data: ... }
    return response.data || response
  }

  const getReceiptByPaymentId = async (paymentId) => {
    const response = await apiCall(`/receipts/payment/${paymentId}`)
    // Single object endpoints return data directly, not wrapped in { data: ... }
    return response.data || response
  }

  const getReceiptPDF = async (receiptId) => {
    const token = getAuthToken()
    const response = await fetch(`${API_BASE_URL}/receipts/${receiptId}/pdf`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })

    if (!response.ok) {
      throw new Error('Failed to download receipt PDF')
    }

    return response.blob()
  }

  return {
    login,
    logout,
    getInvoices,
    getInvoiceById,
    createInvoice,
    updateInvoice,
    issueInvoice,
    getPayments,
    getPaymentById,
    createPayment,
    cancelPayment,
    createPaymentNote,
    getFeeTypes,
    createFeeType,
    updateFeeType,
    getStudents,
    getAccountsReport,
    getReportSummary,
    exportFinanceReport,
    getFinanceDashboard,
    getProgrammes,
    getSemesters,
    getUsers,
    createUser,
    getUserById,
    updateUser,
    getAuditLogs,
    getReceipt,
    getReceiptByPaymentId,
    getReceiptPDF
  }
}

// Utility functions
export function formatRWF(amount) {
  return `RWF ${new Intl.NumberFormat('en-RW', { maximumFractionDigits: 0 }).format(Number(amount) || 0)}`
}

export function formatDate(value, options = { day: '2-digit', month: 'short', year: 'numeric' }) {
  if (!value) return '—'
  const date = new Date(value.length === 10 ? `${value}T00:00:00` : value)
  return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat('en-GB', options).format(date)
}

export function escapeHTML(value = '') {
  return String(value).replace(/[&<>"']/g, character => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;'
  })[character])
}

export function statusBadge(status) {
  const safeStatus = escapeHTML(status)
  return `<span class="status-pill status-${safeStatus.toLowerCase().replace(/[^a-z]+/g, '-')}">${safeStatus}</span>`
}
