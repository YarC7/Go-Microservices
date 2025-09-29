import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Products from './pages/Products';
import Orders from './pages/Orders';

function App() {
  return (
    <Router>
      <div className="App">
        <Layout>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/products" element={<Products />} />
            <Route path="/orders" element={<Orders />} />
            <Route path="/inventory" element={<div className="text-center py-12"><h2 className="text-xl text-gray-600">Inventory Management - Coming Soon</h2></div>} />
            <Route path="/payments" element={<div className="text-center py-12"><h2 className="text-xl text-gray-600">Payment Management - Coming Soon</h2></div>} />
            <Route path="/notifications" element={<div className="text-center py-12"><h2 className="text-xl text-gray-600">Notifications - Coming Soon</h2></div>} />
            <Route path="/analytics" element={<div className="text-center py-12"><h2 className="text-xl text-gray-600">Analytics Dashboard - Coming Soon</h2></div>} />
          </Routes>
        </Layout>
        
        <ToastContainer
          position="top-right"
          autoClose={3000}
          hideProgressBar={false}
          newestOnTop={false}
          closeOnClick
          rtl={false}
          pauseOnFocusLoss
          draggable
          pauseOnHover
          theme="light"
        />
      </div>
    </Router>
  )
}

export default App
