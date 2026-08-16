// components/Card.js - PREMIUM STYLE
import { motion } from 'framer-motion';

export default function Card({ card, faceDown = false, selected = false, layoutId }) {
  if (!card && !faceDown) return <div className="w-[100px] h-[150px] sm:w-[120px] sm:h-[180px] md:w-36 md:h-52 lg:w-48 lg:h-72 bg-white/5 rounded-2xl border-2 border-dashed border-white/10 animate-pulse" />;

  return (
    <motion.div
      layoutId={layoutId}
      initial={false}
      animate={{ rotateY: faceDown ? 180 : 0, scale: 1 }}
      exit={{ opacity: 0, scale: 0.8 }}
      transition={{ duration: 0.6, type: "spring", stiffness: 260, damping: 20 }}
      style={{ transformStyle: 'preserve-3d' }}
      className={`relative w-[100px] h-[150px] sm:w-[120px] sm:h-[180px] md:w-36 md:h-52 lg:w-48 lg:h-72 rounded-2xl ${selected ? 'aura-pulse scale-105 z-10' : 'shadow-xl'}`}
    >
      {/* Front Face (faceDown = false) */}
      <div 
        className={`absolute inset-0 rounded-2xl overflow-hidden backdrop-blur-md flex flex-col transition-all duration-300 ${selected ? 'border-[3px] border-blue-400' : 'border border-white/10 bg-black/40'}`}
        style={{ backfaceVisibility: 'hidden', transform: 'rotateY(0deg)' }}
      >
        {card && (
          <>
            {/* Image Section */}
            <div className="h-[50%] w-full relative bg-slate-900">
              <img src={card.image} alt={card.name} className="w-full h-full object-cover object-top" />
              <div className="absolute bottom-0 left-0 w-full h-12 bg-gradient-to-t from-black/80 to-transparent" />
            </div>

            {/* Stats Section */}
            <div className="flex-1 bg-black/60 p-2 lg:p-3 flex flex-col justify-between border-t border-white/10 relative z-20 overflow-hidden">
              <div className="font-bold text-white text-xs sm:text-base md:text-lg mb-1 truncate drop-shadow-md">{card.name}</div>
              
              <div className="grid grid-cols-2 gap-x-1 lg:gap-x-2 gap-y-1 sm:gap-y-2 overflow-hidden">
                {Object.entries(card.stats).map(([k, v]) => (
                  <div key={k} className="flex flex-col justify-center min-w-0">
                    <div className="flex justify-between items-center gap-0.5 lg:gap-1 min-w-0">
                      <div className="text-slate-400 uppercase tracking-wider font-bold text-[5px] sm:text-[7px] md:text-[8px] lg:text-[10px] truncate shrink">{k}</div>
                      <div className="font-mono font-bold text-white text-[6px] sm:text-[9px] md:text-[10px] lg:text-xs shrink-0">{Math.floor(v)}</div>
                    </div>
                    {/* Stat Bar */}
                    <div className="w-full h-0.5 sm:h-1 bg-white/10 rounded-full mt-0.5 overflow-hidden">
                      <motion.div initial={{ width: 0 }} animate={{ width: `${Math.floor(v)}%` }} transition={{ duration: 1 }} className={`h-full ${selected ? 'bg-yellow-400' : 'bg-purple-500'}`} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
      </div>

      {/* Back Face (faceDown = true) */}
      <div 
        className={`absolute inset-0 rounded-2xl overflow-hidden backdrop-blur-md flex flex-col border-2 border-yellow-600/30 bg-slate-900 shadow-xl shadow-black/50`}
        style={{ backfaceVisibility: 'hidden', transform: 'rotateY(180deg)' }}
      >
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-slate-800 via-black to-black opacity-90" />
        {/* Geometric Pattern */}
        <div className="absolute inset-2 border-2 border-yellow-600/40 rounded-xl" />
        <div className="absolute inset-4 border border-yellow-500/20 rounded-lg" />
        <div className="absolute inset-0 flex items-center justify-center">
            <div className="w-16 h-24 md:w-24 md:h-32 border-4 border-yellow-600/60 rotate-45 absolute" />
            <div className="w-16 h-24 md:w-24 md:h-32 border-4 border-yellow-600/60 -rotate-45 absolute" />
            <div className="w-10 h-10 md:w-16 md:h-16 bg-yellow-500/20 rounded-full blur-xl absolute" />
            <h2 className="text-xl md:text-2xl font-black text-transparent bg-clip-text bg-gradient-to-br from-yellow-300 to-yellow-600 drop-shadow-lg z-10 tracking-widest text-center">META<br/>CLASH</h2>
        </div>
      </div>
    </motion.div>
  );
}