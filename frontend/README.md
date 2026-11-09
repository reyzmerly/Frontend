# University of Kigali - Invoice & Payment Management Frontend

Nuxt 3 frontend for the University of Kigali invoice and payment management system.

## Prerequisites

- Node.js 22.22.3 or higher
- npm

## Installation

```bash
npm install
```

## Development

To run in development mode:

```bash
npm run dev
```

The app will run on port 8080 by default.

## Production Build

To build for production:

```bash
npm run build
```

To run the production build:

```bash
./start.sh
```

Or manually:

```bash
PORT=8080 node .output/server/index.mjs
```

## Features

- **Student Portal**: View invoices, make payments, access payment history and receipts
- **Finance Portal**: Manage invoices, fee types, monitor payments, generate reports
- **Admin Portal**: Manage user roles and access, view audit logs
- **Authentication**: JWT-based authentication with role-based access control
- **API Integration**: Full integration with Go backend (http://localhost:9107/api/v1)

## Running

The frontend runs on **http://localhost:8080**

## Backend Integration

The frontend connects to the Go backend running on `http://localhost:9107/api/v1`.

### Test Accounts

- **Student**: student@uk.ac.rw / password123
- **Finance**: finance@uk.ac.rw / password123
- **Admin**: admin@uk.ac.rw / password123

## Project Structure

- `app.vue` - Root layout
- `pages/index.vue` - Main application page with login and workspace
- `components/` - Vue components for each role and page
- `composables/useApi.js` - API integration composable
- `assets/css/` - Original CSS styles preserved from vanilla JS version
- `public/` - Static assets (images)

## Design

The frontend preserves the original University of Kigali branding and design:
- DM Sans and Manrope fonts
- University of Kigali burgundy/gold/green color scheme
- Responsive design for mobile and desktop
- Dashboard cards, tables, and status indicators
