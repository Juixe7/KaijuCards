// frontend/pages/profile.js - PREMIUM GLASS UI
import { useEffect, useState } from 'react';
import { motion } from 'framer-motion';
import { api } from '../lib/api';

export default function Profile() {
  const [profile, setProfile] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchProfile = async () => {
      const token = localStorage.getItem('meta_clash_token');
      const userId = localStorage.getItem('meta_clash_user_id');

      if (!token || !userId) {
        window.location.href = '/login';
        return;
      }

      try {
        const data = await api.get(`/api/users/${userId}`);
        setProfile(data);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('meta_clash_token');
    localStorage.removeItem('meta_clash_user_id');
    localStorage.removeItem('lastPlayerName');
    window.location.href = '/login';
  };

  if (loading) return <div className="min-h-screen bg-premium flex items-center justify-center text-white font-bold tracking-widest animate-pulse">LOADING...</div>;

  return (
    <div className="min-h-screen relative p-8 overflow-hidden font-sans text-white">
      <div className="bg-premium" />
      <div className="bg-orb w-96 h-96 bg-fuchsia-600 top-20 right-20 opacity-20" />
      <div className="bg-orb w-80 h-80 bg-blue-600 bottom-20 left-20 animation-delay-2000 opacity-20" />

      <div className="max-w-5xl mx-auto relative z-10">
        <header className="flex justify-between items-center mb-12">
          <h1 className="text-3xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 to-fuchsia-400">
            <a href="/">META CLASH</a>
          </h1>
          <button onClick={handleLogout} className="px-6 py-2 rounded-full border border-red-500/30 text-red-400 hover:bg-red-500/10 font-bold tracking-widest text-sm transition-all">
            LOGOUT
          </button>
        </header>

        {error ? (
          <div className="bg-red-500/20 text-red-200 p-4 rounded-xl text-center">{error}</div>
        ) : profile ? (
          <div className="grid md:grid-cols-3 gap-8">
            
            {/* Identity Card */}
            <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="glass-panel p-8 rounded-3xl md:col-span-1 flex flex-col items-center text-center">
              <div className="w-32 h-32 rounded-full bg-gradient-to-br from-indigo-500 to-fuchsia-600 flex items-center justify-center text-5xl font-black shadow-[0_0_30px_rgba(139,92,246,0.3)] mb-6">
                {profile.username[0].toUpperCase()}
              </div>
              <h2 className="text-3xl font-bold mb-1">{profile.username}</h2>
              <div className="text-xs font-mono text-slate-400 mb-8 tracking-widest">ID: {profile.id.split('-')[0]}</div>

              <div className="w-full grid grid-cols-2 gap-4">
                <div className="bg-white/5 rounded-2xl p-4 border border-white/5">
                  <div className="text-xs text-slate-400 uppercase tracking-widest mb-1">Wins</div>
                  <div className="text-3xl font-black text-emerald-400">{profile.wins}</div>
                </div>
                <div className="bg-white/5 rounded-2xl p-4 border border-white/5">
                  <div className="text-xs text-slate-400 uppercase tracking-widest mb-1">Losses</div>
                  <div className="text-3xl font-black text-red-400">{profile.losses}</div>
                </div>
              </div>
            </motion.div>

            {/* Match History */}
            <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="glass-panel p-8 rounded-3xl md:col-span-2 flex flex-col">
              <h3 className="text-xl font-bold uppercase tracking-widest text-slate-300 mb-6">Combat History</h3>
              
              <div className="flex-1 space-y-3 overflow-y-auto custom-scrollbar pr-2 max-h-[500px]">
                {profile.history && profile.history.length > 0 ? (
                  profile.history.map((match, i) => (
                    <div key={i} className="flex items-center justify-between p-4 bg-white/5 border border-white/5 rounded-xl hover:bg-white/10 transition-colors">
                      <div className="flex items-center gap-4">
                        <div className={`w-12 h-12 rounded-lg flex items-center justify-center font-bold text-lg ${match.won ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-red-500/20 text-red-400 border border-red-500/30'}`}>
                          {match.won ? 'W' : 'L'}
                        </div>
                        <div>
                          <div className="font-bold text-lg capitalize">{match.theme}</div>
                          <div className="text-xs font-mono text-slate-400">{new Date(match.finishedAt).toLocaleString()}</div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-xs text-slate-400 uppercase tracking-widest mb-1">Score</div>
                        <div className="font-black text-xl">{match.score} <span className="text-slate-500 text-sm">/ 6</span></div>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="h-full flex flex-col items-center justify-center text-slate-500 opacity-50">
                    <div className="text-4xl mb-4">⚔️</div>
                    <div className="uppercase tracking-widest text-sm font-bold">No battles fought yet</div>
                  </div>
                )}
              </div>
            </motion.div>

          </div>
        ) : null}
      </div>
    </div>
  );
}
