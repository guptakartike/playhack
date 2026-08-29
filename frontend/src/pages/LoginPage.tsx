import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

type Step = 'role' | 'signin';
type Role = 'student' | 'admin';

export default function LoginPage() {
  const navigate = useNavigate();
  const { requestOtp, verifyOtp } = useAuth();

  const [step, setStep] = useState<Step>('role');
  const [role, setRole] = useState<Role>('student');
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [otpSent, setOtpSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [serverOtpCode, setServerOtpCode] = useState<string | null>(null);

  const handleRoleSelect = (selectedRole: Role) => {
    setRole(selectedRole);
    setStep('signin');
    setError(null);
  };

  const handleRequestOtp = async () => {
    if (!email) return;
    setError(null);
    setLoading(true);

    try {
      const code = await requestOtp(email.trim());
      setOtpSent(true);
      if (code) {
        setServerOtpCode(code);
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to send OTP. Please ensure email ends with @iitg.ac.in');
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOtp = async () => {
    if (!otp) return;
    setError(null);
    setLoading(true);

    try {
      await verifyOtp(email.trim(), otp.trim());
      navigate('/sports');
    } catch (err: any) {
      setError(err?.message || 'Invalid or expired OTP. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-dvh bg-blush flex flex-col">
      {/* Top section with logo */}
      <div className="flex-1 flex flex-col items-center justify-center px-8">
        {/* Logo */}
        <div className="flex flex-col items-center mb-10">
          <div className="w-24 h-24 rounded-3xl bg-white flex items-center justify-center shadow-lg shadow-plum/10 p-3.5 mb-4 border border-blush-dark/30">
            <img src="/icon.svg" alt="Huddle Up Logo" className="w-full h-full object-contain" />
          </div>
          <h1 className="text-3xl font-extrabold text-plum tracking-tight">Huddle Up</h1>
          <p className="text-sm text-plum-muted mt-1">Book your game. Own the court.</p>
        </div>

        {step === 'role' ? (
          /* ── Role Selection ── */
          <div className="w-full max-w-sm flex flex-col gap-3">
            <button
              onClick={() => handleRoleSelect('student')}
              className="w-full bg-plum text-white py-4 rounded-2xl font-semibold text-base hover:bg-plum-light transition-all duration-200 active:scale-[0.98] shadow-lg shadow-plum/15"
            >
              Student Login
            </button>
            <button
              onClick={() => handleRoleSelect('admin')}
              className="w-full bg-white text-plum py-4 rounded-2xl font-semibold text-base border-2 border-plum/15 hover:border-plum/30 hover:bg-blush-light transition-all duration-200 active:scale-[0.98]"
            >
              Admin Login
            </button>
          </div>
        ) : (
          /* ── Sign In Form ── */
          <div className="w-full max-w-sm">
            <div className="flex items-center gap-3 mb-6">
              <button
                onClick={() => {
                  setStep('role');
                  setOtpSent(false);
                  setEmail('');
                  setOtp('');
                  setError(null);
                  setServerOtpCode(null);
                }}
                className="w-9 h-9 flex items-center justify-center rounded-full hover:bg-blush-dark/40 transition-colors"
              >
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#271F30" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M19 12H5" />
                  <path d="M12 19l-7-7 7-7" />
                </svg>
              </button>
              <h2 className="text-xl font-bold text-plum">
                Sign In
                <span className="text-sm font-medium text-plum-muted ml-2 capitalize">({role})</span>
              </h2>
            </div>

            {error && (
              <div className="mb-4 bg-danger/10 border border-danger/20 rounded-xl p-3 flex items-start gap-2">
                <span className="text-danger text-xs mt-0.5">⚠️</span>
                <p className="text-xs font-medium text-danger">{error}</p>
              </div>
            )}

            <div className="flex flex-col gap-4">
              {/* Email Input */}
              <div>
                <label className="text-xs font-semibold text-plum-muted tracking-wide uppercase mb-1.5 block">
                  College Email (@iitg.ac.in)
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder={role === 'student' ? 'student@iitg.ac.in' : 'admin@iitg.ac.in'}
                  disabled={otpSent || loading}
                  className="w-full bg-white rounded-xl px-4 py-3.5 text-sm text-plum placeholder:text-plum-muted/50 border-2 border-transparent focus:border-plum/20 focus:outline-none transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                />
              </div>

              {!otpSent ? (
                <button
                  onClick={handleRequestOtp}
                  disabled={!email || loading}
                  className="w-full bg-plum text-white py-3.5 rounded-xl font-semibold text-sm hover:bg-plum-light transition-all duration-200 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {loading ? (
                    <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  ) : (
                    'Send OTP'
                  )}
                </button>
              ) : (
                <>
                  {/* OTP sent confirmation */}
                  <div className="flex flex-col gap-1.5 bg-success/10 border border-success/20 rounded-xl px-4 py-2.5">
                    <div className="flex items-center gap-2">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#22C55E" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="20 6 9 17 4 12" />
                      </svg>
                      <span className="text-xs font-medium text-success">OTP sent to {email}</span>
                    </div>
                    {serverOtpCode && (
                      <div className="flex items-center justify-between pt-1 border-t border-success/20">
                        <span className="text-[11px] text-success/80">Received code:</span>
                        <button
                          type="button"
                          onClick={() => setOtp(serverOtpCode)}
                          className="text-xs font-bold text-success bg-white px-2 py-0.5 rounded shadow-xs hover:scale-105 transition-transform tracking-wider"
                        >
                          {serverOtpCode} (Click to auto-fill)
                        </button>
                      </div>
                    )}
                  </div>

                  {/* OTP Input */}
                  <div>
                    <label className="text-xs font-semibold text-plum-muted tracking-wide uppercase mb-1.5 block">
                      6-Digit OTP
                    </label>
                    <input
                      type="text"
                      value={otp}
                      onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                      placeholder="••••••"
                      maxLength={6}
                      inputMode="numeric"
                      autoFocus
                      disabled={loading}
                      className="w-full bg-white rounded-xl px-4 py-3.5 text-sm text-plum placeholder:text-plum-muted/50 border-2 border-transparent focus:border-plum/20 focus:outline-none transition-colors tracking-[0.4em] font-bold text-center text-lg"
                    />
                  </div>

                  <button
                    onClick={handleVerifyOtp}
                    disabled={otp.length < 6 || loading}
                    className="w-full bg-plum text-white py-3.5 rounded-xl font-semibold text-sm hover:bg-plum-light transition-all duration-200 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                  >
                    {loading ? (
                      <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    ) : (
                      <>
                        Verify & Continue
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                          <line x1="5" y1="12" x2="19" y2="12" />
                          <polyline points="12 5 19 12 12 19" />
                        </svg>
                      </>
                    )}
                  </button>

                  <button
                    type="button"
                    onClick={() => {
                      setOtpSent(false);
                      setOtp('');
                      setServerOtpCode(null);
                      setError(null);
                    }}
                    className="text-xs font-medium text-plum-muted hover:text-plum transition-colors text-center"
                  >
                    Resend OTP or Change Email
                  </button>
                </>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Bottom tagline */}
      <div className="pb-[max(2rem,env(safe-area-inset-bottom))] text-center">
        <p className="text-[10px] text-plum-muted/50 tracking-wide">IIT Guwahati • Campus Sports Booking</p>
      </div>
    </div>
  );
}
