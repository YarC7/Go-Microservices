import axios from 'axios';

// API Base URL - using API Gateway
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL 
  ? `${import.meta.env.VITE_API_BASE_URL}/api/v1` 
  : '/api/v1';

// Configure axios
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Types
export interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
}

export interface Order {
  id: number;
  customer_id: number;
  product_id: number;
  quantity: number;
  total_price: number;
  status: string;
  created_at?: string;
}

export interface OrderWithPayment extends Order {
  currency: string;
}

export interface InventoryItem {
  id: number;
  product_id: number;
  quantity: number;
  sku: string;
  location?: string;
}

export interface Notification {
  id: number;
  order_id: number;
  customer_id: number;
  message: string;
  status: string;
  created_at: string;
  delivered_at?: string;
}

export interface Payment {
  id: number;
  order_id: number;
  customer_id: number;
  amount: number;
  currency: string;
  status: string;
  stripe_payment_id: string;
  stripe_client_secret?: string;
  payment_method?: string;
  created_at: string;
  updated_at: string;
}

export interface PaymentResponse {
  payment: Payment;
  client_secret?: string;
  message?: string;
}

// Product Service API
export const productService = {
  getAll: async (): Promise<Product[]> => {
    const response = await api.get('/products');
    return response.data;
  },
  
  getById: async (id: number): Promise<Product> => {
    const response = await api.get(`/products/${id}`);
    return response.data;
  },
  
  create: async (product: Omit<Product, 'id'>): Promise<Product> => {
    const response = await api.post('/products', product);
    return response.data;
  },
  
  update: async (id: number, product: Partial<Product>): Promise<Product> => {
    const response = await api.put(`/products/${id}`, product);
    return response.data;
  },
  
  delete: async (id: number): Promise<void> => {
    await api.delete(`/products/${id}`);
  },
};

// Order Service API
export const orderService = {
  getAll: async (): Promise<Order[]> => {
    const response = await api.get('/orders');
    return response.data;
  },
  
  getById: async (id: number): Promise<Order> => {
    const response = await api.get(`/orders/${id}`);
    return response.data;
  },
  
  create: async (order: Omit<Order, 'id'>): Promise<Order> => {
    const response = await api.post('/orders', order);
    return response.data;
  },
  
  createWithPayment: async (order: Omit<OrderWithPayment, 'id'>): Promise<{
    order: Order;
    payment: PaymentResponse;
  }> => {
    const response = await api.post('/orders/with-payment', order);
    return response.data;
  },
  
  createBatch: async (orders: Omit<Order, 'id'>[]): Promise<{
    total_orders: number;
    successful: number;
    failed: number;
    failed_orders: any[];
    processing_time: string;
  }> => {
    const response = await api.post('/orders/batch', orders);
    return response.data;
  },
  
  update: async (id: number, order: Partial<Order>): Promise<Order> => {
    const response = await api.put(`/orders/${id}`, order);
    return response.data;
  },
  
  updateStatus: async (id: number, status: string): Promise<{ message: string; order_id: number; status: string }> => {
    const response = await api.patch(`/orders/${id}/status`, { status });
    return response.data;
  },
  
  delete: async (id: number): Promise<void> => {
    await api.delete(`/orders/${id}`);
  },
};

// Inventory Service API
export const inventoryService = {
  getAll: async (): Promise<InventoryItem[]> => {
    const response = await api.get('/inventory');
    return response.data;
  },
  
  getById: async (id: number): Promise<InventoryItem> => {
    const response = await api.get(`/inventory/${id}`);
    return response.data;
  },
  
  checkAvailability: async (productId: number, quantity: number): Promise<{ available: boolean }> => {
    const response = await api.post('/inventory/check', { product_id: productId, quantity });
    return response.data;
  },
  
  create: async (item: Omit<InventoryItem, 'id'>): Promise<InventoryItem> => {
    const response = await api.post('/inventory', item);
    return response.data;
  },
  
  update: async (id: number, item: Partial<InventoryItem>): Promise<InventoryItem> => {
    const response = await api.put(`/inventory/${id}`, item);
    return response.data;
  },
  
  delete: async (id: number): Promise<void> => {
    await api.delete(`/inventory/${id}`);
  },
};

// Notification Service API
export const notificationService = {
  getAll: async (): Promise<Notification[]> => {
    const response = await api.get('/notifications');
    return response.data;
  },
  
  getById: async (id: number): Promise<Notification> => {
    const response = await api.get(`/notifications/${id}`);
    return response.data;
  },
  
  getByCustomer: async (customerId: number): Promise<Notification[]> => {
    const response = await api.get(`/notifications/customer/${customerId}`);
    return response.data;
  },
  
  create: async (notification: Omit<Notification, 'id' | 'created_at'>): Promise<Notification> => {
    const response = await api.post('/notifications', notification);
    return response.data;
  },
  
  markAsDelivered: async (id: number): Promise<Notification> => {
    const response = await api.put(`/notifications/${id}/deliver`);
    return response.data;
  },
};

// Payment Service API
export const paymentService = {
  create: async (payment: {
    order_id: number;
    customer_id: number;
    amount: number;
    currency: string;
  }): Promise<PaymentResponse> => {
    const response = await api.post('/payments', payment);
    return response.data;
  },
  
  confirm: async (paymentIntentId: string): Promise<PaymentResponse> => {
    const response = await api.post('/payments/confirm', { payment_intent_id: paymentIntentId });
    return response.data;
  },
  
  getById: async (id: number): Promise<Payment> => {
    const response = await api.get(`/payments/${id}`);
    return response.data;
  },
  
  getByOrder: async (orderId: number): Promise<Payment[]> => {
    const response = await api.get(`/payments/order/${orderId}`);
    return response.data;
  },
};

// Health Check
export const healthService = {
  check: async (): Promise<{ status: string }> => {
    const response = await api.get('/health');
    return response.data;
  },
};

export default api;