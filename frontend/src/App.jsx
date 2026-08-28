import React, { useState, useEffect } from 'react';
import { Trophy, CalendarDays, LogIn, LogOut, Sparkles, CheckCircle, Flame } from 'lucide-react';
import MyBookingsPage from './pages/MyBookingsPage';
import { getStoredToken, setStoredToken, clearStoredToken } from './api/client';

export default function App() {
  const [token, setToken] = useState(getStoredToken());
  const [activePage, setActivePage] = useState(token ? 'my-bookings' : 'login');
  
  // Login State
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [step, setStep] = useState('request'); // 'request' | 'verify'
  const [otpSentCode, setOtpSentCode] = useState('');
  const [loginError, setLoginError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

  useEffect(() => {
    if (!token && activePage === 'my-bookings') {
      setActivePage('login');
    }
  }, [token, activePage]);

  const handleRequestOTP = async (e, customEmail) => {
    if (e) e.preventDefault();
    const targetEmail = customEmail || email;

    if (!targetEmail.endsWith('@iitg.ac.in')) {
      setLoginError('Email must end with @iitg.ac.in');
      return;
    }

    setLoginError('');
    setIsSubmitting(true);

    try {
      const res = await fetch(`${API_BASE_URL}/auth/request-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: targetEmail }),
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Failed to send OTP');
      }

      setEmail(targetEmail);
      setOtpSentCode(data.code || '');
      setOtp(data.code || ''); // Auto-fill for quick hackathon testing
      setStep('verify');
    } catch (err) {
      setLoginError(err.message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleVerifyOTP = async (e) => {
    if (e) e.preventDefault();
    setLoginError('');
    setIsSubmitting(true);

    try {
      const res = await fetch(`${API_BASE_URL}/auth/verify-otp`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, code: otp }),
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || 'Failed to verify OTP');
      }

      setStoredToken(data.token);
      setToken(data.token);
      setActivePage('my-bookings');
    } catch (err) {
      setLoginError(err.message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleLogout = () => {
    clearStoredToken();
    setToken('');
    setStep('request');
    setEmail('');
    setOtp('');
    setActivePage('login');
  };

  return (
    <div className="min-h-screen bg-[#0B0B12] text-[#F1F5F9] font-sans flex flex-col">
      {/* Navigation Header */}
      <header className="sticky top-0 z-40 bg-[#0B0B12]/80 backdrop-blur-md border-b border-[#1F1F2E]">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 h-16 flex items-center justify-between">
          {/* Logo */}
          <div 
            onClick={() => setActivePage(token ? 'my-bookings' : 'login')}
            className="flex items-center gap-2 cursor-pointer group"
          >
            <div className="p-2 rounded-xl bg-gradient-to-tr from-[#F5793A] to-orange-400 text-white shadow-lg shadow-[#F5793A]/20">
              <Flame className="w-5 h-5 fill-white" />
            </div>
            <span className="text-lg font-black tracking-wider uppercase text-white">
              PLAY <span className="text-[#F5793A]">HACK</span>
            </span>
          </div>

          {/* Nav Items */}
          <nav className="flex items-center gap-2">
            {token ? (
              <>
                <button
                  onClick={() => setActivePage('my-bookings')}
                  className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold transition-all ${
                    activePage === 'my-bookings'
                      ? 'bg-[#F5793A]/10 text-[#F5793A] border border-[#F5793A]/30'
                      : 'text-[#94A3B8] hover:text-white hover:bg-[#14141E]'
                  }`}
                >
                  <CalendarDays className="w-4 h-4" />
                  <span>My Bookings</span>
                </button>

                <button
                  onClick={handleLogout}
                  className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold text-rose-400 hover:bg-rose-950/20 border border-transparent hover:border-rose-900/40 transition-all"
                >
                  <LogOut className="w-4 h-4" />
                  <span className="hidden sm:inline">Logout</span>
                </button>
              </>
            ) : (
              <button
                onClick={() => setActivePage('login')}
                className="flex items-center gap-1.5 px-4 py-2 rounded-xl text-xs font-bold bg-[#F5793A] hover:bg-[#E06728] text-white shadow-md shadow-[#F5793A]/20 transition-all"
              >
                <LogIn className="w-4 h-4" />
                <span>Login</span>
              </button>
            )}
          </nav>
        </div>
      </header>

      {/* Main Container */}
      <main className="flex-1">
        {activePage === 'my-bookings' && token ? (
          <MyBookingsPage
            onNavigateBrowse={() => alert('Facility Browse module active!')}
            onNavigateLogin={handleLogout}
          />
        ) : (
          /* Login Screen */
          <div className="max-w-md mx-auto px-4 py-16">
            <div className="bg-[#14141E] border border-[#2A2A3C] rounded-3xl p-6 sm:p-8 shadow-2xl space-y-6">
              <div className="text-center space-y-2">
                <div className="w-12 h-12 rounded-2xl bg-[#F5793A]/10 border border-[#F5793A]/30 flex items-center justify-center mx-auto text-[#F5793A]">
                  <Trophy className="w-6 h-6" />
                </div>
                <h2 className="text-2xl font-black tracking-tight text-[#F1F5F9]">
                  Welcome to PlayHack
                </h2>
                <p className="text-xs text-[#94A3B8]">
                  Domain-restricted college email login (@iitg.ac.in)
                </p>
              </div>

              {/* Quick Demo Login Buttons */}
              <div className="p-3 rounded-2xl bg-[#1A1A26] border border-[#2A2A3C] space-y-2">
                <span className="text-[11px] font-semibold uppercase tracking-wider text-[#94A3B8] flex items-center gap-1">
                  <Sparkles className="w-3 h-3 text-[#F5793A]" /> Quick Demo Autofill
                </span>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => handleRequestOTP(null, 'test@iitg.ac.in')}
                    className="flex-1 py-1.5 px-3 rounded-xl bg-[#2A2A3C] hover:bg-[#F5793A]/20 hover:border-[#F5793A]/40 border border-transparent text-xs text-[#F1F5F9] font-medium transition-all text-center"
                  >
                    test@iitg.ac.in
                  </button>
                  <button
                    type="button"
                    onClick={() => handleRequestOTP(null, 'judge1@iitg.ac.in')}
                    className="flex-1 py-1.5 px-3 rounded-xl bg-[#2A2A3C] hover:bg-[#F5793A]/20 hover:border-[#F5793A]/40 border border-transparent text-xs text-[#F1F5F9] font-medium transition-all text-center"
                  >
                    judge1@iitg.ac.in
                  </button>
                </div>
              </div>

              {loginError && (
                <div className="p-3 rounded-xl bg-rose-950/40 border border-rose-900/50 text-xs text-rose-300 text-center font-medium">
                  {loginError}
                </div>
              )}

              {step === 'request' ? (
                <form onSubmit={handleRequestOTP} className="space-y-4">
                  <div className="space-y-1.5">
                    <label className="text-xs font-semibold text-[#94A3B8]">College Email</label>
                    <input
                      type="email"
                      required
                      placeholder="student@iitg.ac.in"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="w-full px-4 py-3 rounded-xl bg-[#1A1A26] border border-[#2A2A3C] text-sm text-[#F1F5F9] focus:outline-none focus:border-[#F5793A] transition-all placeholder:text-slate-600"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={isSubmitting}
                    className="w-full py-3 rounded-xl text-xs font-bold bg-[#F5793A] hover:bg-[#E06728] text-white shadow-lg shadow-[#F5793A]/20 transition-all disabled:opacity-50"
                  >
                    {isSubmitting ? 'Sending OTP...' : 'Send OTP Code'}
                  </button>
                </form>
              ) : (
                <form onSubmit={handleVerifyOTP} className="space-y-4">
                  {otpSentCode && (
                    <div className="p-3 rounded-xl bg-emerald-950/30 border border-emerald-900/50 text-xs text-emerald-300 flex items-center justify-between">
                      <span>Prototype Code: <strong className="font-mono text-white">{otpSentCode}</strong></span>
                      <CheckCircle className="w-4 h-4 text-emerald-400" />
                    </div>
                  )}

                  <div className="space-y-1.5">
                    <label className="text-xs font-semibold text-[#94A3B8]">6-Digit OTP</label>
                    <input
                      type="text"
                      required
                      placeholder="123456"
                      value={otp}
                      onChange={(e) => setOtp(e.target.value)}
                      className="w-full px-4 py-3 rounded-xl bg-[#1A1A26] border border-[#2A2A3C] text-center font-mono tracking-widest text-lg text-[#F1F5F9] focus:outline-none focus:border-[#F5793A] transition-all"
                    />
                  </div>

                  <div className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setStep('request')}
                      className="px-4 py-3 rounded-xl text-xs font-medium text-[#94A3B8] hover:text-white bg-[#1A1A26] transition-colors"
                    >
                      Back
                    </button>
                    <button
                      type="submit"
                      disabled={isSubmitting}
                      className="flex-1 py-3 rounded-xl text-xs font-bold bg-[#F5793A] hover:bg-[#E06728] text-white shadow-lg shadow-[#F5793A]/20 transition-all disabled:opacity-50"
                    >
                      {isSubmitting ? 'Verifying...' : 'Verify & Login'}
                    </button>
                  </div>
                </form>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
