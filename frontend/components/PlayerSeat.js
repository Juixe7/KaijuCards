// components/PlayerSeat.js
import Card from './Card';

export default function PlayerSeat({ player, position, isTurn }) {
  if (!player) return null;

  const flexDir = {
    bottom: 'flex-row',
    top: 'flex-row-reverse',
    left: 'flex-col-reverse',
    right: 'flex-col-reverse'
  }[position] || 'flex-col';

  // Fallback to array length if hand is hidden (remote players might just have a  length)
  const cardCount = player.hand ? player.hand.length : 0;
  
  const originClass = {
    bottom: 'origin-bottom',
    top: 'origin-top',
    left: 'origin-left',
    right: 'origin-right'
  }[position] || 'origin-center';

  return (
    <div className="absolute z-10" style={getSeatStyles(position)}>
      <div className={`flex items-center justify-center gap-4 ${flexDir} transition-all duration-500 scale-[0.65] sm:scale-75 lg:scale-100 ${originClass}`}>
        
        {/* Profile Container */}
        <div className={`p-3 rounded-2xl border transition-all duration-300 relative overflow-hidden backdrop-blur-xl w-32 sm:w-40 md:w-48 ${isTurn ? 'bg-indigo-900/80 border-indigo-400 aura-pulse scale-110 z-20' : 'bg-black/60 border-white/10'}`}>
          {isTurn && <div className="absolute inset-0 bg-radial-gradient from-purple-500/30 to-transparent" />}
          <div className="flex justify-between items-center mb-2 relative z-10">
            <div className="font-bold flex items-center gap-1.5 text-xs md:text-sm text-white truncate">
               {player.name}
               {player.isBot && <span className="text-[8px] bg-yellow-500/20 text-yellow-400 px-1.5 py-0.5 rounded border border-yellow-500/20">BOT</span>}
            </div>
            <div className="text-base md:text-lg font-black text-indigo-300">{player.score || 0}</div>
          </div>
          <div className="h-1 bg-black/50 rounded-full overflow-hidden relative z-10">
            <div style={{ width: `${((player.score || 0) / 6) * 100}%` }} className="h-full bg-gradient-to-r from-indigo-500 to-purple-500 transition-all duration-500" />
          </div>
        </div>

        {/* Hand Stack (Face down) */}
        <div className={`relative mt-2 md:mt-0 ${(position === 'left' || position === 'right') ? '-translate-x-4 md:-translate-x-6' : ''}`} style={{ width: '80px', height: '120px' }}>
           {Array.from({ length: cardCount }).map((_, idx) => {
             // Provide a consistent layoutId so we can animate from hand to table.
             const cardData = player.hand?.[idx];
             const layoutId = cardData?.id ? `card-${cardData.id}` : `card-hidden-${player.id}-${idx}`;
             const fanDir = position === 'top' ? -12 : 12;
             
             return (
               <div key={idx} className="absolute" style={{
                  top: 0,
                  left: idx * fanDir,
                  zIndex: idx,
                  transform: 'scale(0.45)',
                  transformOrigin: 'top left'
               }}>
                 <div className="pointer-events-none">
                   <Card card={cardData} faceDown={true} layoutId={layoutId} />
                 </div>
               </div>
             );
           })}
        </div>
      </div>
    </div>
  );
}

function getSeatStyles(position) {
  // We use fixed absolute positioning around the screen.
  // Responsive tweaks can be added via classes, but inline styles guarantee the anchor points.
  switch (position) {
    case 'bottom': return { bottom: '10px', left: '50%', transform: 'translateX(-50%)' };
    case 'top': return { top: '10px', left: '50%', transform: 'translateX(-50%)' };
    case 'left': return { left: '10px', top: '50%', transform: 'translateY(-50%)' };
    case 'right': return { right: '10px', top: '50%', transform: 'translateY(-50%)' };
    default: return {};
  }
}
