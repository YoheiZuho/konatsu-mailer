// SPDX-License-Identifier: Apache-2.0

import React from 'react';
import ReactDOM from 'react-dom/client';
import { QueryClientProvider } from '@tanstack/react-query';
import { BrowserRouter } from 'react-router-dom';

import './index.css';
import { App } from './App';
import { queryClient } from '@/lib/queryClient';
import { initAppearance } from '@/stores/appearance';
import { registerServiceWorker } from '@/lib/pwa';

// Apply persisted theme/brand to the DOM as early as possible.
initAppearance();
registerServiceWorker();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  </React.StrictMode>,
);
