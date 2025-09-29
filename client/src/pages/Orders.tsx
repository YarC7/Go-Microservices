import React, { useState, useEffect } from 'react';
import { Plus, Eye, Edit, Trash2, Search, CreditCard, Clock, CheckCircle, XCircle } from 'lucide-react';
import { orderService, productService, inventoryService, paymentService } from '../services/api';
import type { Order, Product, Payment } from '../services/api';
import { toast } from 'react-toastify';

const Orders: React.FC = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [payments, setPayments] = useState<{ [orderId: number]: Payment[] }>({});
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [showModal, setShowModal] = useState(false);
  const [modalType, setModalType] = useState<'create' | 'view' | 'edit'>('create');
  const [selectedOrder, setSelectedOrder] = useState<Order | null>(null);
  const [formData, setFormData] = useState({
    customer_id: 1,
    product_id: '',
    quantity: 1,
    total_price: 0,
    currency: 'USD',
    with_payment: false,
  });

  useEffect(() => {
    fetchOrders();
    fetchProducts();
  }, []);

  const fetchOrders = async () => {
    try {
      const data = await orderService.getAll();
      setOrders(data);
      
      // Fetch payments for each order
      const paymentPromises = data.map(async (order) => {
        try {
          const orderPayments = await paymentService.getByOrder(order.id);
          return { orderId: order.id, payments: orderPayments };
        } catch (error) {
          return { orderId: order.id, payments: [] };
        }
      });
      
      const paymentResults = await Promise.all(paymentPromises);
      const paymentsMap: { [orderId: number]: Payment[] } = {};
      paymentResults.forEach(({ orderId, payments }) => {
        paymentsMap[orderId] = payments;
      });
      setPayments(paymentsMap);
      
    } catch (error) {
      console.error('Error fetching orders:', error);
      toast.error('Failed to fetch orders');
    } finally {
      setLoading(false);
    }
  };

  const fetchProducts = async () => {
    try {
      const data = await productService.getAll();
      setProducts(data);
    } catch (error) {
      console.error('Error fetching products:', error);
    }
  };

  const getProductName = (productId: number): string => {
    const product = products.find(p => p.id === productId);
    return product ? product.name : `Product ${productId}`;
  };

  const getProductPrice = (productId: number): number => {
    const product = products.find(p => p.id === productId);
    return product ? product.price : 0;
  };

  const filteredOrders = orders.filter(order => {
    const matchesSearch = 
      order.id.toString().includes(searchTerm) ||
      getProductName(order.product_id).toLowerCase().includes(searchTerm.toLowerCase()) ||
      order.status.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesStatus = statusFilter === 'all' || order.status === statusFilter;
    
    return matchesSearch && matchesStatus;
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const productPrice = getProductPrice(parseInt(formData.product_id));
      const totalPrice = productPrice * formData.quantity;
      
      const orderData = {
        customer_id: formData.customer_id,
        product_id: parseInt(formData.product_id),
        quantity: formData.quantity,
        total_price: totalPrice,
      };

      let result;
      if (formData.with_payment) {
        // Create order with payment
        result = await orderService.createWithPayment({
          ...orderData,
          currency: formData.currency,
        });
        
        setOrders([...orders, result.order]);
        
        // Update payments state
        setPayments(prev => ({
          ...prev,
          [result.order.id]: [result.payment.payment]
        }));
        
        toast.success('Order created with payment intent');
      } else {
        // Create regular order
        const newOrder = await orderService.create(orderData);
        setOrders([...orders, newOrder]);
        toast.success('Order created successfully');
      }
      
      resetForm();
    } catch (error) {
      console.error('Error creating order:', error);
      toast.error('Failed to create order');
    }
  };

  const handleStatusUpdate = async (orderId: number, newStatus: string) => {
    try {
      await orderService.updateStatus(orderId, newStatus);
      setOrders(orders.map(order => 
        order.id === orderId ? { ...order, status: newStatus } : order
      ));
      toast.success('Order status updated');
    } catch (error) {
      console.error('Error updating order status:', error);
      toast.error('Failed to update order status');
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this order?')) {
      return;
    }

    try {
      await orderService.delete(id);
      setOrders(orders.filter(o => o.id !== id));
      toast.success('Order deleted successfully');
    } catch (error) {
      console.error('Error deleting order:', error);
      toast.error('Failed to delete order');
    }
  };

  const resetForm = () => {
    setFormData({
      customer_id: 1,
      product_id: '',
      quantity: 1,
      total_price: 0,
      currency: 'USD',
      with_payment: false,
    });
    setSelectedOrder(null);
    setShowModal(false);
  };

  const openModal = (type: 'create' | 'view' | 'edit', order?: Order) => {
    setModalType(type);
    if (order) {
      setSelectedOrder(order);
      if (type === 'edit') {
        setFormData({
          customer_id: order.customer_id,
          product_id: order.product_id.toString(),
          quantity: order.quantity,
          total_price: order.total_price,
          currency: 'USD',
          with_payment: false,
        });
      }
    }
    setShowModal(true);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case 'pending':
        return <Clock className="h-4 w-4 text-yellow-600" />;
      case 'cancelled':
        return <XCircle className="h-4 w-4 text-red-600" />;
      default:
        return <Clock className="h-4 w-4 text-gray-600" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800';
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'cancelled':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Orders</h1>
          <p className="text-gray-600 mt-2">Manage customer orders and payments</p>
        </div>
        <button
          onClick={() => openModal('create')}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center"
        >
          <Plus className="h-5 w-5 mr-2" />
          New Order
        </button>
      </div>

      {/* Filters */}
      <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-3 h-5 w-5 text-gray-400" />
            <input
              type="text"
              placeholder="Search orders..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
        </div>
      </div>

      {/* Orders Table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Orders List</h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Order ID
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Product
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Customer
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Quantity
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Total
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Payment
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredOrders.map((order) => {
                const orderPayments = payments[order.id] || [];
                const hasPayment = orderPayments.length > 0;
                const paymentStatus = hasPayment ? orderPayments[0].status : 'none';
                
                return (
                  <tr key={order.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      #{order.id}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {getProductName(order.product_id)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      Customer {order.customer_id}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {order.quantity}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      ${order.total_price.toFixed(2)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(order.status)}`}>
                        {getStatusIcon(order.status)}
                        <span className="ml-1">{order.status}</span>
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {hasPayment ? (
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          paymentStatus === 'succeeded' ? 'bg-green-100 text-green-800' :
                          paymentStatus === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                          'bg-red-100 text-red-800'
                        }`}>
                          <CreditCard className="h-3 w-3 mr-1" />
                          {paymentStatus}
                        </span>
                      ) : (
                        <span className="text-xs text-gray-500">No payment</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      <div className="flex space-x-2">
                        <button
                          onClick={() => openModal('view', order)}
                          className="text-blue-600 hover:text-blue-800 transition-colors"
                        >
                          <Eye className="h-4 w-4" />
                        </button>
                        <select
                          value={order.status}
                          onChange={(e) => handleStatusUpdate(order.id, e.target.value)}
                          className="text-xs border border-gray-300 rounded px-2 py-1"
                        >
                          <option value="pending">Pending</option>
                          <option value="completed">Completed</option>
                          <option value="cancelled">Cancelled</option>
                        </select>
                        <button
                          onClick={() => handleDelete(order.id)}
                          className="text-red-600 hover:text-red-800 transition-colors"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        {filteredOrders.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">
              {searchTerm || statusFilter !== 'all' 
                ? 'No orders found matching your criteria.' 
                : 'No orders found. Create your first order!'
              }
            </p>
          </div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-md w-full p-6">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">
              {modalType === 'create' && 'Create New Order'}
              {modalType === 'view' && `Order #${selectedOrder?.id}`}
              {modalType === 'edit' && 'Edit Order'}
            </h2>
            
            {modalType === 'view' && selectedOrder ? (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Order ID</label>
                    <p className="text-sm text-gray-900">#{selectedOrder.id}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Customer</label>
                    <p className="text-sm text-gray-900">Customer {selectedOrder.customer_id}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Product</label>
                    <p className="text-sm text-gray-900">{getProductName(selectedOrder.product_id)}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Quantity</label>
                    <p className="text-sm text-gray-900">{selectedOrder.quantity}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Total Price</label>
                    <p className="text-sm text-gray-900 font-medium">${selectedOrder.total_price.toFixed(2)}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Status</label>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getStatusColor(selectedOrder.status)}`}>
                      {getStatusIcon(selectedOrder.status)}
                      <span className="ml-1">{selectedOrder.status}</span>
                    </span>
                  </div>
                </div>
                
                {/* Payment Information */}
                {payments[selectedOrder.id] && payments[selectedOrder.id].length > 0 && (
                  <div className="mt-6 pt-4 border-t border-gray-200">
                    <h3 className="text-sm font-medium text-gray-700 mb-3">Payment Information</h3>
                    {payments[selectedOrder.id].map((payment) => (
                      <div key={payment.id} className="bg-gray-50 p-3 rounded-lg">
                        <div className="grid grid-cols-2 gap-2 text-sm">
                          <div>
                            <span className="font-medium">Amount:</span> ${payment.amount.toFixed(2)}
                          </div>
                          <div>
                            <span className="font-medium">Currency:</span> {payment.currency.toUpperCase()}
                          </div>
                          <div>
                            <span className="font-medium">Status:</span> 
                            <span className={`ml-1 px-2 py-0.5 rounded text-xs ${
                              payment.status === 'succeeded' ? 'bg-green-100 text-green-800' :
                              payment.status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                              'bg-red-100 text-red-800'
                            }`}>
                              {payment.status}
                            </span>
                          </div>
                          <div>
                            <span className="font-medium">Method:</span> {payment.payment_method || 'N/A'}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                <div className="flex justify-end pt-4">
                  <button
                    onClick={resetForm}
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                  >
                    Close
                  </button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleSubmit} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Product
                  </label>
                  <select
                    required
                    value={formData.product_id}
                    onChange={(e) => setFormData({ ...formData, product_id: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value="">Select a product</option>
                    {products.map((product) => (
                      <option key={product.id} value={product.id}>
                        {product.name} - ${product.price.toFixed(2)}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Quantity
                  </label>
                  <input
                    type="number"
                    required
                    min="1"
                    value={formData.quantity}
                    onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) || 1 })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                </div>

                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={formData.with_payment}
                      onChange={(e) => setFormData({ ...formData, with_payment: e.target.checked })}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-sm font-medium text-gray-700">Create with payment intent</span>
                  </label>
                </div>

                {formData.with_payment && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Currency
                    </label>
                    <select
                      value={formData.currency}
                      onChange={(e) => setFormData({ ...formData, currency: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    >
                      <option value="USD">USD</option>
                      <option value="EUR">EUR</option>
                      <option value="GBP">GBP</option>
                    </select>
                  </div>
                )}

                {formData.product_id && (
                  <div className="bg-gray-50 p-3 rounded-lg">
                    <p className="text-sm text-gray-600">
                      Total: ${(getProductPrice(parseInt(formData.product_id)) * formData.quantity).toFixed(2)}
                    </p>
                  </div>
                )}

                <div className="flex justify-end space-x-3 pt-4">
                  <button
                    type="button"
                    onClick={resetForm}
                    className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center"
                  >
                    {formData.with_payment && <CreditCard className="h-4 w-4 mr-2" />}
                    Create Order
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default Orders;