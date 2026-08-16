import React, { useEffect, useState } from 'react';
import { motion, useMotionValue, useTransform, useSpring } from 'framer-motion';

export default function UniverseForge() {
  const [particles, setParticles] = useState([]);
  
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);

  useEffect(() => {
    // Center initially
    mouseX.set(window.innerWidth / 2);
    mouseY.set(window.innerHeight / 2);

    const handleMouseMove = (e) => {
      mouseX.set(e.clientX);
      mouseY.set(e.clientY);
    };
    window.addEventListener('mousemove', handleMouseMove);

    // Generate random data fragments
    const newParticles = Array.from({ length: 40 }).map((_, i) => {
      const angle = Math.random() * Math.PI * 2;
      const radius = 120 + Math.random() * 200; // start 120px to 320px away
      const startX = Math.cos(angle) * radius;
      const startY = Math.sin(angle) * radius;
      const delay = Math.random() * 2; // 0 to 2s delay
      const duration = 1 + Math.random(); // 1 to 2s duration
      return { id: i, startX, startY, delay, duration };
    });
    setParticles(newParticles);

    return () => window.removeEventListener('mousemove', handleMouseMove);
  }, [mouseX, mouseY]);

  // Map mouse to parallax offset
  const rawOffsetX = useTransform(mouseX, [0, typeof window !== 'undefined' ? window.innerWidth : 1000], [-30, 30]);
  const rawOffsetY = useTransform(mouseY, [0, typeof window !== 'undefined' ? window.innerHeight : 1000], [-30, 30]);

  // Apply spring physics for buttery smooth transition
  const smoothX = useSpring(rawOffsetX, { stiffness: 50, damping: 20 });
  const smoothY = useSpring(rawOffsetY, { stiffness: 50, damping: 20 });

  return (
    <div className="relative flex items-center justify-center w-64 h-64 mb-8 perspective-1000">
      
      {/* Holographic Orbital Rings */}
      <motion.div 
        animate={{ rotateZ: 360, rotateX: 75, rotateY: 15 }}
        transition={{ duration: 8, repeat: Infinity, ease: "linear" }}
        className="absolute w-48 h-48 border-2 border-indigo-500/40 rounded-full shadow-[0_0_15px_rgba(99,102,241,0.3)]"
      />
      <motion.div 
        animate={{ rotateZ: -360, rotateX: -60, rotateY: -45 }}
        transition={{ duration: 12, repeat: Infinity, ease: "linear" }}
        className="absolute w-56 h-56 border border-fuchsia-500/30 rounded-full shadow-[0_0_20px_rgba(217,70,239,0.2)]"
      />
      <motion.div 
        animate={{ rotateZ: 180, rotateX: 45, rotateY: 75 }}
        transition={{ duration: 15, repeat: Infinity, ease: "linear" }}
        className="absolute w-64 h-64 border border-cyan-500/20 rounded-full border-dashed"
      />

      {/* Parallax Group (Core + Particles) */}
      <motion.div 
        style={{ x: smoothX, y: smoothY }}
        className="absolute inset-0 flex items-center justify-center"
      >
        {/* The Central Universe Core */}
        <motion.div 
          animate={{ rotate: 45 }}
          className="absolute z-10 w-16 h-16 bg-gradient-to-br from-indigo-400 via-fuchsia-500 to-pink-500 shadow-[0_0_50px_rgba(217,70,239,0.8)] border border-white/50 flex items-center justify-center overflow-hidden"
        >
          {/* Nano-tech inner glow */}
          <div className="absolute inset-2 border border-white/30 bg-black/20" />
          {/* Holographic Scanlines */}
          <div className="absolute inset-0 bg-[linear-gradient(transparent_50%,rgba(0,0,0,0.5)_50%)] bg-[length:100%_4px] opacity-40 animate-pulse" />
        </motion.div>

        {/* Nano-tech Assembling Particles */}
        <div className="absolute inset-0 pointer-events-none">
          {particles.map(p => (
            <motion.div
              key={p.id}
              initial={{ x: p.startX, y: p.startY, opacity: 0, scale: 0 }}
              animate={{ 
                x: [p.startX, 0, 0], 
                y: [p.startY, 0, 0], 
                opacity: [0, 1, 0],
                scale: [0, 1.5, 0.2]
              }}
              transition={{
                duration: p.duration,
                repeat: Infinity,
                delay: p.delay,
                ease: "circIn"
              }}
              className="absolute top-1/2 left-1/2 w-1.5 h-1.5 bg-fuchsia-200 shadow-[0_0_10px_rgba(255,255,255,1)] rounded-full -ml-[3px] -mt-[3px]"
            />
          ))}
        </div>
      </motion.div>
      
    </div>
  );
}
