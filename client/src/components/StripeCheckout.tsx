import React, { useState, useEffect } from 'react';
import { loadStripe, StripeElementsOptions } from '@stripe/stripe-js';
import {
  Elements,
  CardElement,
  useStripe,
  useElements
} from '@stripe/react-stripe-js';
import { CreditCard, Lock } from 'lucide-react';
import { paymentService } from '../services/api';
import { toast } from 'react-toastify';

// Load Stripe with your publishable key
const STRIPE_PUBLISHABLE_KEY = import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY || 'pk_test_51234567890abcdef...';
const stripePromise = loadStripe(STRIPE_PUBLISHABLE_KEY);

interface StripeCheckoutProps {
  orderId: number;
  customerId: number;
  amount: number;
  currency: string;
  onSuccess: (paymentIntent: any) => void;
  onCancel: () => void;
}

const CheckoutForm: React.FC<StripeCheckoutProps> = ({ orderId, customerId, amount, currency, onSuccess, onCancel }) => {
  const stripe = useStripe();
  const elements = useElements();
  const [loading, setLoading] = useState(false);
  const [clientSecret, setClientSecret] = useState<string | null>(null);

  useEffect(() => {
    createPaymentIntent();
  }, []);

  const createPaymentIntent = async () => {
    try {
      const response = await paymentService.create({
        order_id: orderId,
        customer_id: customerId,
        amount: amount,
        currency: currency,
      });
      
      if (response.client_secret) {
        setClientSecret(response.client_secret);
      }
    } catch (error) {
      console.error('Error creating payment intent:', error);
      toast.error('Failed to initialize payment');
    }
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();

    if (!stripe || !elements || !clientSecret) {
      return;
    }

    setLoading(true);

    const cardElement = elements.getElement(CardElement);
    if (!cardElement) {
      setLoading(false);
      return;
    }

    const { error, paymentIntent } = await stripe.confirmCardPayment(
      clientSecret,
      {
        payment_method: {
          card: cardElement,
          billing_details: {
            name: `Customer ${customerId}`,
          },
        },
      }
    );

    setLoading(false);

    if (error) {
      console.error('Payment failed:', error);
      toast.error(error.message || 'Payment failed');
    } else {
      toast.success('Payment successful!');
      onSuccess(paymentIntent);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="bg-gray-50 p-4 rounded-lg">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Payment Details</h3>
          <Lock className="h-5 w-5 text-green-600" />
        </div>
        
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Card Information
            </label>
            <div className="p-3 border border-gray-300 rounded-lg bg-white">
              <CardElement />
            </div>
          </div>
        </div>
      </div>

      <div className="bg-white p-4 rounded-lg border border-gray-200">
        <h4 className="font-medium text-gray-900 mb-3">Order Summary</h4>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-gray-600">Order #{orderId}</span>
            <span className="text-gray-900">Customer {customerId}</span>
          </div>
          <div className="flex justify-between font-medium">
            <span className="text-gray-900">Total:</span>
            <span className="text-gray-900">${amount.toFixed(2)} {currency.toUpperCase()}</span>
          </div>
        </div>
      </div>

      <div className="flex space-x-3">
        <button
          type="button"
          onClick={onCancel}
          className="flex-1 px-4 py-3 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors font-medium"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={!stripe || loading || !clientSecret}
          className="flex-1 px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 transition-colors font-medium flex items-center justify-center"
        >
          {loading ? (
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
          ) : (
            <>
              <CreditCard className="h-5 w-5 mr-2" />
              Pay ${amount.toFixed(2)}
            </>
          )}
        </button>
      </div>

      <div className="text-xs text-gray-500 text-center">
        <p>Test: 4242 4242 4242 4242, any expiry/CVC</p>
      </div>
    </form>
  );
};

const StripeCheckout: React.FC<StripeCheckoutProps> = (props) => {
  const options: StripeElementsOptions = {
    appearance: {
      theme: 'stripe',
      variables: {
        colorPrimary: '#2563eb',
      },
    },
  };

  return (
    <div className="max-w-md mx-auto">
      <Elements stripe={stripePromise} options={options}>
        <CheckoutForm {...props} />
      </Elements>
    </div>
  );
};

export default StripeCheckout;
