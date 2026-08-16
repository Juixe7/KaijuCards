// frontend/pages/register.js - PREMIUM GLASS AUTH UI
import { useState } from 'react';
import { motion } from 'framer-motion';
import { api } from '../lib/api';

export default function Register() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleRegister = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      await api.post('/api/auth/register', { username, password });

      // Automatically redirect to login after successful registration
      window.location.href = '/login';
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen relative flex items-center justify-center p-4 overflow-hidden font-sans">
      <div className="bg-premium" />
      <div className="bg-orb w-96 h-96 bg-emerald-600 top-10 right-10 opacity-30" />
      <div className="bg-orb w-80 h-80 bg-teal-600 bottom-20 left-20 animation-delay-2000 opacity-30" />

      <motion.div 
        initial={{ opacity: 0, scale: 0.95 }} 
        animate={{ opacity: 1, scale: 1 }} 
        className="glass-panel w-full max-w-md rounded-3xl p-8 relative z-10"
      >
        <div className="text-center mb-8">
          <h1 className="text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-emerald-400 to-teal-400">REGISTER</h1>
          <p className="text-slate-400 mt-2 font-mono text-xs tracking-widest uppercase">Establish Identity</p>
        </div>

        {error && (
          <div className="bg-red-500/20 border border-red-500/50 text-red-200 p-3 rounded-lg text-sm mb-6 text-center font-bold tracking-wide">
            {error}
          </div>
        )}

        <form onSubmit={handleRegister} className="space-y-6">
          <div className="space-y-2">
            <label className="text-xs font-bold text-emerald-300 uppercase tracking-widest ml-1">Username</label>
            <input 
              value={username} 
              onChange={e => setUsername(e.target.value)} 
              placeholder="3-30 characters" 
              required
              minLength={3}
              maxLength={30}
              className="glass-input w-full p-4 rounded-xl text-lg font-medium placeholder-slate-600 focus:ring-2 focus:ring-emerald-500/50" 
            />
          </div>
          
          <div className="space-y-2">
            <label className="text-xs font-bold text-teal-300 uppercase tracking-widest ml-1">Password</label>
            <input 
              type="password"
              value={password} 
              onChange={e => setPassword(e.target.value)} 
              placeholder="Minimum 6 characters" 
              required
              minLength={6}
              className="glass-input w-full p-4 rounded-xl text-lg font-medium placeholder-slate-600 focus:ring-2 focus:ring-teal-500/50" 
            />
          </div>

          <button 
            type="submit" 
            disabled={loading}
            className={`w-full p-4 rounded-xl font-bold text-xl tracking-widest uppercase transition-all duration-300 ${loading ? 'opacity-50 cursor-not-allowed bg-slate-800' : 'bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white hover:scale-[1.02] shadow-[0_0_20px_rgba(16,185,129,0.3)]'}`}
          >
            {loading ? 'Processing...' : 'Create Account'}
          </button>
        </form>

        <div className="mt-8 text-center">
          <p className="text-slate-400 text-sm">
            Already exist? <a href="/login" className="text-emerald-400 font-bold hover:text-emerald-300 hover:underline transition-colors">Login Here</a>
          </p>
        </div>
      </motion.div>
    </div>
  );
}
