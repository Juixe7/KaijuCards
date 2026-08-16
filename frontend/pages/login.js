// frontend/pages/login.js - PREMIUM GLASS AUTH UI
import { useState } from 'react';
import { motion } from 'framer-motion';
import { api } from '../lib/api';

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const data = await api.post('/api/auth/login', { username, password });

      localStorage.setItem('meta_clash_token', data.token);
      localStorage.setItem('meta_clash_user_id', data.user.id);
      localStorage.setItem('lastPlayerName', data.user.username);
      
      window.location.href = '/';
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen relative flex items-center justify-center p-4 overflow-hidden font-sans">
      <div className="bg-premium" />
      <div className="bg-orb w-96 h-96 bg-purple-600 top-10 left-10 opacity-30" />
      <div className="bg-orb w-80 h-80 bg-blue-600 bottom-20 right-20 animation-delay-2000 opacity-30" />

      <motion.div 
        initial={{ opacity: 0, y: 20 }} 
        animate={{ opacity: 1, y: 0 }} 
        className="glass-panel w-full max-w-md rounded-3xl p-8 relative z-10"
      >
        <div className="text-center mb-8">
          <h1 className="text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-purple-400">LOGIN</h1>
          <p className="text-slate-400 mt-2 font-mono text-xs tracking-widest uppercase">Authenticate Identity</p>
        </div>

        {error && (
          <div className="bg-red-500/20 border border-red-500/50 text-red-200 p-3 rounded-lg text-sm mb-6 text-center font-bold tracking-wide">
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-6">
          <div className="space-y-2">
            <label className="text-xs font-bold text-indigo-300 uppercase tracking-widest ml-1">Username</label>
            <input 
              value={username} 
              onChange={e => setUsername(e.target.value)} 
              placeholder="Enter Username" 
              required
              className="glass-input w-full p-4 rounded-xl text-lg font-medium placeholder-slate-600 focus:ring-2 focus:ring-indigo-500/50" 
            />
          </div>
          
          <div className="space-y-2">
            <label className="text-xs font-bold text-purple-300 uppercase tracking-widest ml-1">Password</label>
            <input 
              type="password"
              value={password} 
              onChange={e => setPassword(e.target.value)} 
              placeholder="Enter Password" 
              required
              className="glass-input w-full p-4 rounded-xl text-lg font-medium placeholder-slate-600 focus:ring-2 focus:ring-purple-500/50" 
            />
          </div>

          <button 
            type="submit" 
            disabled={loading}
            className={`w-full p-4 rounded-xl font-bold text-xl tracking-widest uppercase transition-all duration-300 ${loading ? 'opacity-50 cursor-not-allowed bg-slate-800' : 'btn-primary hover:scale-[1.02]'}`}
          >
            {loading ? 'Authenticating...' : 'Enter System'}
          </button>
        </form>

        <div className="mt-8 text-center">
          <p className="text-slate-400 text-sm">
            No identity? <a href="/register" className="text-indigo-400 font-bold hover:text-indigo-300 hover:underline transition-colors">Register Here</a>
          </p>
        </div>
      </motion.div>
    </div>
  );
}
