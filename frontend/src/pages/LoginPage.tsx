import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

type Step = 'role' | 'signin';
type Role = 'student' | 'admin';

export default function LoginPage() {
  const navigate = useNavigate();
  const [step, setStep] = useState<Step>('role');
  const [role, setRole] = useState<Role>('student');
  const [email, setEmail] = useState('');
  const [otp, setOtp] = useState('');
  const [otpSent, setOtpSent] = useState(false);
  const [loading, setLoading] = useState(false);

  const handleRoleSelect = (selectedRole: Role) => {
    setRole(selectedRole);
    setStep('signin');
  };

  const handleRequestOtp = () => {
    if (!email) return;
    setLoading(true);
    // Simulate OTP request
    setTimeout(() => {
      setOtpSent(true);
      setLoading(false);
    }, 800);
  };

  const handleVerifyOtp = () => {
    if (!otp) return;
    setLoading(true);
    // Simulate verification
    setTimeout(() => {
      setLoading(false);
      navigate('/sports');
    }, 800);
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
            <div className="flex items-center gap-3 mb-8">
              <button
                onClick={() => { setStep('role'); setOtpSent(false); setEmail(''); setOtp(''); }}
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

            <div className="flex flex-col gap-4">
              {/* Email Input */}
              <div>
                <label className="text-xs font-semibold text-plum-muted tracking-wide uppercase mb-1.5 block">
                  Email
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder={role === 'student' ? 'name@iitg.ac.in' : 'admin@iitg.ac.in'}
                  disabled={otpSent}
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
                  <div className="flex items-center gap-2 bg-success/10 rounded-xl px-4 py-2.5">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#22C55E" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                      <polyline points="20 6 9 17 4 12" />
                    </svg>
                    <span className="text-xs font-medium text-success">OTP sent to {email}</span>
                  </div>

                  {/* OTP Input */}
                  <div>
                    <label className="text-xs font-semibold text-plum-muted tracking-wide uppercase mb-1.5 block">
                      OTP
                    </label>
                    <input
                      type="text"
                      value={otp}
                      onChange={(e) => setOtp(e.target.value.replace(/\D/g, '').slice(0, 6))}
                      placeholder="6-digit code"
                      maxLength={6}
                      inputMode="numeric"
                      autoFocus
                      className="w-full bg-white rounded-xl px-4 py-3.5 text-sm text-plum placeholder:text-plum-muted/50 border-2 border-transparent focus:border-plum/20 focus:outline-none transition-colors tracking-[0.3em] font-semibold text-center text-lg"
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
                    onClick={() => { setOtpSent(false); setOtp(''); }}
                    className="text-xs font-medium text-plum-muted hover:text-plum transition-colors text-center"
                  >
                    Resend OTP
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
