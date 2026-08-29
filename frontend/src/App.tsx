import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { NotificationProvider } from './context/NotificationContext';
import LoginPage from './pages/LoginPage';
import BookCourtPage from './pages/BookCourtPage';
import CourtDetailsPage from './pages/CourtDetailsPage';
import SelectTimePage from './pages/SelectTimePage';
import CheckoutPage from './pages/CheckoutPage';
import MyBookingsPage from './pages/MyBookingsPage';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <NotificationProvider>
          <Routes>
            <Route path="/" element={<LoginPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/sports" element={<BookCourtPage />} />
            <Route path="/book" element={<BookCourtPage />} />
            <Route path="/courts/:sport" element={<CourtDetailsPage />} />
            <Route path="/book/:sport/court/:courtId/time" element={<SelectTimePage />} />
            <Route path="/checkout/:sport/:courtId/:time" element={<CheckoutPage />} />
            <Route path="/bookings" element={<MyBookingsPage />} />
            <Route path="/profile" element={<MyBookingsPage />} />
          </Routes>
        </NotificationProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}
