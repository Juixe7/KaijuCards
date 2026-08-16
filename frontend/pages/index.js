// pages/index.js - PREMIUM GLASS LOBBY
import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/router';
import wsClient from '../lib/ws';
import { motion, AnimatePresence } from 'framer-motion';
import UniverseForge from '../components/UniverseForge';

export default function Home() {
  const router = useRouter();
  const [lobby, setLobby] = useState(null);
  const [name, setName] = useState('');
  const [theme, setTheme] = useState('One Piece');
  const [loadingGame, setLoadingGame] = useState(false);
  const [loadingFact, setLoadingFact] = useState('');
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const containerRef = useRef(null);

  const handleMouseMove = (e) => {
    if (containerRef.current) {
      containerRef.current.style.setProperty('--mouse-x', `${e.clientX}px`);
      containerRef.current.style.setProperty('--mouse-y', `${e.clientY}px`);
    }
  };

  const facts = [
    "Rule #1: The highest attribute always wins the round. Choose your stats wisely.",
    "Lore: The Meta Clash Engine weaves infinite universes from a single theme.",
    "Pro Tip: Every universe has its own unique strengths. Adapt your strategy.",
    "Did you know? Meta Clash generates unique, balanced cards on the fly using advanced AI.",
    "Lore: Across the multiverse, champions rise and fall. Will you be a legend?",
    "Rule #2: Winning rounds deals damage to your opponent. Drain their health to claim victory!"
  ];


  useEffect(() => {
    if (typeof window !== 'undefined') {
      setName(localStorage.getItem('lastPlayerName') || 'Player');
      setTheme(localStorage.getItem('lastTheme') || 'One Piece');
      setIsLoggedIn(!!localStorage.getItem('meta_clash_token'));
    }

    wsClient.connect();
    
    const onLobbyUpdate = (l) => { setLobby(l); };
    const onGameStarted = (l) => {
      setLobby(l);
      localStorage.setItem('lastLobbyId', l.id);
      setLoadingFact(facts[Math.floor(Math.random() * facts.length)]);
      setLoadingGame(true);
      setTimeout(() => {
        router.push('/game');
      }, 10000);
    };

    wsClient.on('lobbyUpdate', onLobbyUpdate);
    wsClient.on('gameStarted', onGameStarted);
    
    return () => {
      wsClient.off('lobbyUpdate', onLobbyUpdate);
      wsClient.off('gameStarted', onGameStarted);
    };
  }, []);

  useEffect(() => { if (typeof window !== 'undefined') localStorage.setItem('lastPlayerName', name); }, [name]);
  useEffect(() => { if (typeof window !== 'undefined') localStorage.setItem('lastTheme', theme); }, [theme]);

  const create = () => wsClient.emit('createLobby', { name, theme }, res => { if (res.ok) setLobby(res.lobby); else alert("Error creating lobby: " + (res.err || "Unknown error")); });
  const join = () => { if (!lobby?.id) return alert('Enter ID'); wsClient.emit('joinLobby', { lobbyId: lobby.id, name }, res => { if (res.ok) setLobby(res.lobby); else alert(res.err); }); };
  const addBot = () => wsClient.emit('addBot', { lobbyId: lobby.id }, res => { if (res.ok) setLobby(res.lobby); });
  const start = () => wsClient.emit('startGame', { lobbyId: lobby.id }, res => { if (!res.ok) alert(res.err); });
  const setLobbyId = (id) => setLobby(l => ({ ...l, id: id }));

  const players = lobby ? lobby.players : [];
  const emptySlots = Array(4 - players.length).fill(null);

  return (
    <div 
      ref={containerRef}
      className="min-h-screen relative flex items-center justify-end p-4 md:pr-12 lg:pr-24 overflow-hidden font-sans"
      onMouseMove={handleMouseMove}
    >
      <AnimatePresence>
        {loadingGame && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-black/95 backdrop-blur-md overflow-hidden"
          >
            <div className="bg-arena pointer-events-none z-0" />
            <div className="hex-grid pointer-events-none z-0" />
            <div className="z-10 flex flex-col items-center justify-center w-full h-full">
              <UniverseForge />
              <h2 className="text-3xl font-black text-white tracking-widest uppercase mb-4 animate-pulse">Forging Universe...</h2>
              <p className="text-indigo-300 font-mono text-center max-w-lg px-4">{loadingFact}</p>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      <div className="bg-synthwave">
        <div className="synthwave-stars" />
        <div className="shooting-stars">
          <span /><span /><span /><span /><span />
        </div>
        <div className="synthwave-sun" />
        <div className="synthwave-horizon" />
        <div className="synthwave-grid" />
      </div>

      <motion.div 
        layout
        initial={{ opacity: 0, scale: 0.95, x: 20 }} 
        animate={{ opacity: 1, scale: 1, x: 0 }} 
        className="retro-panel p-6 md:p-10 relative z-10 flex flex-col md:flex-row gap-8 items-start"
      >
        <motion.div layout className="flex flex-col gap-8 w-full md:w-[420px] shrink-0">
          <div className="flex flex-col gap-2 border-b border-indigo-500/20 pb-6">
            <div className="flex justify-between items-start">
              <h1 className="text-4xl md:text-5xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 tracking-tight leading-none uppercase">
                META<br/>CLASH
              </h1>
              <div className="flex flex-col items-end gap-2">
                {isLoggedIn ? (
                  <a href="/profile" className="px-3 py-1.5 bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 rounded-full font-bold tracking-widest text-[10px] hover:bg-indigo-500/30 transition-all uppercase">
                    My Profile
                  </a>
                ) : (
                  <div className="flex flex-col gap-2">
                    <a href="/login" className="px-3 py-1.5 bg-white/5 text-white border border-white/10 rounded-full font-bold tracking-widest text-[10px] hover:bg-white/10 transition-all uppercase text-center">
                      Login
                    </a>
                    <a href="/register" className="px-3 py-1.5 bg-indigo-600 text-white rounded-full font-bold tracking-widest text-[10px] hover:bg-indigo-500 transition-all uppercase shadow-lg shadow-indigo-500/30 text-center">
                      Register
                    </a>
                  </div>
                )}
              </div>
            </div>
            <p className="text-indigo-400/80 text-sm font-mono tracking-widest uppercase">Generate. Battle. Conquer.</p>
          </div>

          <div className="space-y-4">
            <div className="space-y-1">
              <label className="text-[10px] font-bold text-indigo-300 uppercase tracking-widest ml-1">Your Identity</label>
              <input value={name} onChange={e => setName(e.target.value)} placeholder="Enter Nickname" className="w-full bg-black/40 border border-indigo-500/30 p-3 rounded-xl text-md font-medium text-white placeholder-slate-600 focus:outline-none focus:border-indigo-400 transition-colors" />
            </div>
            
            <div className="space-y-1">
              <label className="text-[10px] font-bold text-purple-300 uppercase tracking-widest ml-1">Universe Theme</label>
              <input value={theme} onChange={e => setTheme(e.target.value)} placeholder="e.g. One Piece, Naruto" className="w-full bg-black/40 border border-purple-500/30 p-3 rounded-xl text-md font-medium text-white placeholder-slate-600 focus:outline-none focus:border-purple-400 transition-colors" />
            </div>
          </div>

          <div className="space-y-3">
             <button onClick={create} disabled={!name || !theme || !!lobby?.id} className={`w-full p-4 rounded-xl font-bold text-lg tracking-widest uppercase transition-all duration-300 border ${!name || !theme || !!lobby?.id ? 'opacity-30 bg-slate-900 border-slate-700 text-slate-500' : 'bg-indigo-600/80 hover:bg-indigo-500 border-indigo-400 text-white shadow-[0_0_20px_rgba(79,70,229,0.4)]'}`}>Create Universe</button>
             <div className="flex gap-2 h-12">
               <input value={lobby?.id || ''} onChange={e => setLobbyId(e.target.value.toUpperCase())} placeholder="LOBBY CODE" className="w-2/3 bg-black/60 px-3 rounded-xl text-center font-mono text-lg font-bold tracking-widest uppercase text-white border border-slate-700 focus:outline-none focus:border-blue-500" />
               <button onClick={join} className="w-1/3 bg-slate-800 hover:bg-slate-700 border border-slate-600 hover:border-slate-500 text-white rounded-xl text-xs font-bold tracking-wider transition-all uppercase">JOIN</button>
             </div>
          </div>
        </motion.div>
        
        <AnimatePresence>
          {lobby?.id && (
            <motion.div 
              layout
              initial={{ opacity: 0, width: 0, paddingLeft: 0, marginLeft: 0 }} 
              animate={{ opacity: 1, width: 320, paddingLeft: 32, marginLeft: 0 }} 
              exit={{ opacity: 0, width: 0, paddingLeft: 0, marginLeft: 0 }}
              className="flex flex-col gap-3 border-l border-emerald-500/30 overflow-hidden shrink-0 h-full min-h-[400px]"
            >
               <div className="flex justify-between items-center mb-2">
                 <div>
                   <div className="text-[9px] text-emerald-400 font-bold uppercase tracking-widest mb-0.5">Active Session</div>
                   <div className="text-xl font-mono font-bold tracking-widest text-white">{lobby.id}</div>
                 </div>
                 <span className="text-[10px] font-mono bg-black/40 px-2 py-1 rounded text-emerald-300 border border-emerald-900">{players.length} / 4 Players</span>
               </div>
               
               <div className="flex flex-col gap-2 flex-1">
                 {players.map((p, i) => (
                   <div key={i} className="flex items-center gap-3 p-2 rounded-lg bg-black/40 border border-white/5">
                     <div className="w-8 h-8 rounded bg-indigo-900/50 flex items-center justify-center font-bold text-sm text-indigo-300 border border-indigo-700/50">{p.name[0]}</div>
                     <div className="font-mono text-sm text-white truncate">{p.name} {p.isBot && <span className="text-[9px] bg-yellow-500/20 text-yellow-300 px-1.5 py-0.5 ml-2 rounded uppercase tracking-wider">BOT</span>}</div>
                   </div>
                 ))}
                 {emptySlots.map((_, i) => (
                   <div key={i} className="flex items-center gap-3 p-2 rounded-lg border border-dashed border-white/10 opacity-30">
                     <div className="w-8 h-8 rounded border border-dashed border-slate-600 flex items-center justify-center text-xs">+</div>
                     <div className="text-[10px] font-mono text-slate-500 uppercase tracking-widest">Awaiting Player</div>
                   </div>
                 ))}
               </div>

               <div className="flex flex-col gap-2 mt-4">
                 <button onClick={addBot} className="w-full py-3 bg-yellow-500/10 hover:bg-yellow-500/20 text-yellow-200 rounded-lg text-[10px] font-bold border border-yellow-500/30 uppercase tracking-widest transition-colors">+ Add Bot</button>
                 <button onClick={start} disabled={lobby.players.length < 2} className="w-full py-4 bg-red-600/80 hover:bg-red-500 text-white rounded-lg text-xs font-bold border border-red-400 uppercase tracking-widest transition-all disabled:opacity-30 disabled:cursor-not-allowed shadow-[0_0_15px_rgba(220,38,38,0.3)]">Start Match</button>
               </div>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </div>
  );
}