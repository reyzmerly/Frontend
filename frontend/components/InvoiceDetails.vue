<template>
  <div class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h2>Invoice Details</h2>
        <button class="icon-button" @click="close" aria-label="Close">×</button>
      </div>
      <div class="modal-body" v-if="invoice">
        <div class="invoice-header">
          <div>
            <p class="eyebrow">INVOICE</p>
            <h1>{{ invoice.invoice_number || invoice.id }}</h1>
            <p>{{ invoice.description }}</p>
          </div>
          <div class="invoice-status">
            <span v-html="statusBadge(invoice.status)"></span>
          </div>
        </div>

        <div class="invoice-meta">
          <div>
            <small>Issued date</small>
            <strong>{{ formatDate(invoice.issued_at?.Time || invoice.created_at) }}</strong>
          </div>
          <div>
            <small>Due date</small>
            <strong>{{ formatDate(invoice.due_date) }}</strong>
          </div>
          <div>
            <small>Programme</small>
            <strong>{{ invoice.programme || '—' }}</strong>
          </div>
          <div>
            <small>Semester</small>
            <strong>{{ invoice.semester || '—' }}</strong>
          </div>
        </div>

        <div v-if="invoice.items && invoice.items.length > 0" class="invoice-items">
          <h3>Invoice Items</h3>
          <div v-for="(item, index) in invoice.items" :key="index" class="item-row">
            <span>{{ item.fee_name }}</span>
            <strong>{{ formatRWF(item.amount) }}</strong>
          </div>
        </div>

        <div class="invoice-totals">
          <div class="total-row">
            <span>Total amount</span>
            <strong>{{ formatRWF(invoice.total_amount) }}</strong>
          </div>
          <div class="total-row">
            <span>Paid amount</span>
            <strong>{{ formatRWF(invoice.paid_amount || 0) }}</strong>
          </div>
          <div class="total-row highlight">
            <span>Outstanding balance</span>
            <strong>{{ formatRWF(invoice.balance || (invoice.total_amount - (invoice.paid_amount || 0))) }}</strong>
          </div>
        </div>

        <div class="modal-actions">
          <button
            v-if="user?.role === 'student' && invoice.balance > 0"
            class="button button-primary"
            @click="initiatePayment"
          >
            Pay now
          </button>
          <button
            v-if="user?.role === 'student' && !invoice.balance"
            class="button button-secondary"
            @click="viewReceipt"
          >
            View receipt
          </button>
          <button class="button button-quiet" @click="close">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useApi, formatRWF, formatDate, statusBadge } from '~/composables/useApi'

const props = defineProps({
  invoice: Object,
  user: Object
})

const emit = defineEmits(['close', 'navigate'])

const close = () => {
  emit('close')
}

const initiatePayment = () => {
  emit('close')
  emit('navigate', 'Payment')
  localStorage.setItem('selectedInvoice', JSON.stringify(props.invoice))
}

const viewReceipt = () => {
  emit('close')
  const invoiceId = props.invoice.invoice_number || props.invoice.id
  localStorage.setItem('selectedInvoiceId', invoiceId)
  emit('navigate', 'Payment History')
}
</script>

<style scoped>
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

.invoice-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
}

.invoice-header .eyebrow {
  font-size: 12px;
  color: #6b7280;
  margin: 0 0 5px 0;
}

.invoice-header h1 {
  margin: 0 0 5px 0;
  font-size: 24px;
}

.invoice-header p {
  margin: 0;
  color: #6b7280;
}

.invoice-meta {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 20px;
  padding: 15px;
  background: #f9fafb;
  border-radius: 6px;
}

@media (min-width: 768px) {
  .invoice-meta {
    grid-template-columns: repeat(4, 1fr);
  }
}

.invoice-meta small {
  display: block;
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 5px;
}

.invoice-meta strong {
  font-size: 16px;
}

.invoice-items {
  margin-bottom: 20px;
  padding: 15px;
  background: #f9fafb;
  border-radius: 6px;
}

.invoice-items h3 {
  margin: 0 0 15px 0;
  font-size: 14px;
  color: #374151;
}

.item-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid #e5e7eb;
}

.item-row:last-child {
  border-bottom: none;
}

.invoice-totals {
  margin-bottom: 20px;
}

.total-row {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid #e5e7eb;
}

.total-row.highlight {
  background: #fef3c7;
  padding: 15px;
  border-radius: 6px;
  margin-top: 10px;
  border-bottom: none;
}

.total-row.highlight strong {
  color: #92400e;
  font-size: 18px;
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
</style>
