import React, { useState, useEffect } from 'react';
import { 
  ShoppingBag, 
  Package, 
  Warehouse, 
  Bell, 
  CreditCard,
  TrendingUp,
  AlertCircle,
  CheckCircle
} from 'lucide-react';
import { orderService, productService, inventoryService, notificationService, paymentService } from '../services/api';
import type { Order, Product, InventoryItem, Notification, Payment } from '../services/api';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState({
    totalOrders: 0,
    totalProducts: 0,
    totalInventoryItems: 0,
    unreadNotifications: 0,
    totalPayments: 0,
    recentOrders: [] as Order[],
    lowStockItems: [] as InventoryItem[],
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        setLoading(true);
        
        // Fetch all data in parallel
        const [
          orders, 
          products, 
          inventory, 
          notifications,
        ] = await Promise.all([
          orderService.getAll().catch(() => []),
          productService.getAll().catch(() => []),
          inventoryService.getAll().catch(() => []),
          notificationService.getAll().catch(() => []),
        ]);

        // Get recent orders (last 5)
        const recentOrders = orders.slice(-5).reverse();
        
        // Get low stock items (quantity < 10)
        const lowStockItems = inventory.filter(item => item.quantity < 10);
        
        // Count unread notifications
        const unreadNotifications = notifications.filter(n => !n.delivered_at).length;

        setStats({
          totalOrders: orders.length,
          totalProducts: products.length,
          totalInventoryItems: inventory.length,
          unreadNotifications,
          totalPayments: 0, // Will be updated when we get payments
          recentOrders,
          lowStockItems,
        });
      } catch (error) {
        console.error('Error fetching dashboard data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, []);

  const StatCard: React.FC<{
    title: string;
    value: number;
    icon: React.ElementType;
    color: string;
    change?: string;
  }> = ({ title, value, icon: Icon, color, change }) => (
    <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-3xl font-bold text-gray-900">{value}</p>
          {change && (
            <p className="text-sm text-green-600 flex items-center mt-1">
              <TrendingUp className="h-4 w-4 mr-1" />
              {change}
            </p>
          )}
        </div>
        <div className={`p-3 rounded-full ${color}`}>
          <Icon className="h-6 w-6 text-white" />
        </div>
      </div>
    </div>
  );

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 mt-2">Welcome to your e-commerce management dashboard</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-6">
        <StatCard
          title="Total Orders"
          value={stats.totalOrders}
          icon={ShoppingBag}
          color="bg-blue-500"
          change="+12% from last month"
        />
        <StatCard
          title="Products"
          value={stats.totalProducts}
          icon={Package}
          color="bg-green-500"
        />
        <StatCard
          title="Inventory Items"
          value={stats.totalInventoryItems}
          icon={Warehouse}
          color="bg-purple-500"
        />
        <StatCard
          title="Notifications"
          value={stats.unreadNotifications}
          icon={Bell}
          color="bg-orange-500"
        />
        <StatCard
          title="Payments"
          value={stats.totalPayments}
          icon={CreditCard}
          color="bg-indigo-500"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Orders */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Orders</h2>
          {stats.recentOrders.length === 0 ? (
            <p className="text-gray-500 text-center py-8">No recent orders</p>
          ) : (
            <div className="space-y-3">
              {stats.recentOrders.map((order) => (
                <div
                  key={order.id}
                  className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                >
                  <div>
                    <p className="font-medium text-gray-900">Order #{order.id}</p>
                    <p className="text-sm text-gray-600">
                      Product {order.product_id} • Qty: {order.quantity}
                    </p>
                  </div>
                  <div className="text-right">
                    <p className="font-medium text-gray-900">${order.total_price}</p>
                    <span
                      className={`inline-flex items-center px-2 py-1 text-xs font-medium rounded-full ${
                        order.status === 'completed'
                          ? 'bg-green-100 text-green-800'
                          : order.status === 'pending'
                          ? 'bg-yellow-100 text-yellow-800'
                          : 'bg-gray-100 text-gray-800'
                      }`}
                    >
                      {order.status === 'completed' && <CheckCircle className="h-3 w-3 mr-1" />}
                      {order.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Low Stock Alerts */}
        <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Low Stock Alerts</h2>
          {stats.lowStockItems.length === 0 ? (
            <div className="text-center py-8">
              <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-3" />
              <p className="text-gray-500">All items are well stocked</p>
            </div>
          ) : (
            <div className="space-y-3">
              {stats.lowStockItems.map((item) => (
                <div
                  key={item.id}
                  className="flex items-center justify-between p-3 bg-red-50 rounded-lg border border-red-200"
                >
                  <div className="flex items-center">
                    <AlertCircle className="h-5 w-5 text-red-500 mr-3" />
                    <div>
                      <p className="font-medium text-gray-900">SKU: {item.sku}</p>
                      <p className="text-sm text-gray-600">Product ID: {item.product_id}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-bold text-red-600">{item.quantity} left</p>
                    <p className="text-xs text-gray-500">{item.location}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <button className="p-4 bg-blue-50 hover:bg-blue-100 rounded-lg border border-blue-200 text-left transition-colors">
            <Package className="h-8 w-8 text-blue-600 mb-2" />
            <h3 className="font-medium text-gray-900">Add Product</h3>
            <p className="text-sm text-gray-600">Create a new product</p>
          </button>
          <button className="p-4 bg-green-50 hover:bg-green-100 rounded-lg border border-green-200 text-left transition-colors">
            <ShoppingBag className="h-8 w-8 text-green-600 mb-2" />
            <h3 className="font-medium text-gray-900">New Order</h3>
            <p className="text-sm text-gray-600">Process an order</p>
          </button>
          <button className="p-4 bg-purple-50 hover:bg-purple-100 rounded-lg border border-purple-200 text-left transition-colors">
            <Warehouse className="h-8 w-8 text-purple-600 mb-2" />
            <h3 className="font-medium text-gray-900">Update Inventory</h3>
            <p className="text-sm text-gray-600">Manage stock levels</p>
          </button>
          <button className="p-4 bg-orange-50 hover:bg-orange-100 rounded-lg border border-orange-200 text-left transition-colors">
            <Bell className="h-8 w-8 text-orange-600 mb-2" />
            <h3 className="font-medium text-gray-900">View Notifications</h3>
            <p className="text-sm text-gray-600">Check recent alerts</p>
          </button>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;