// lib/game.js - FINAL STABLE (ESM + No Nanoid)
const fetch = (...args) => import('node-fetch').then(({default: f})=>f(...args));
import crypto from 'crypto';
import _ from 'lodash';

// ATTRS and CONSTANTS
export const ATTRS = ['rank','strength','speed','iq'];
export const CARDS_PER_PLAYER = 6;
export const MAX_PLAYERS = 4;
export const TOTAL_CARDS = CARDS_PER_PLAYER * MAX_PLAYERS;

// Helper to generate IDs without external libraries
function generateId() {
    return crypto.randomBytes(4).toString('hex').toUpperCase();
}

async function generatePlausibleStats(name, theme){
    // Simulate AI generation with clean math
    const statData = await new Promise(resolve => {
        const stats = {};
        ATTRS.forEach(attr => {
            const base = (name.length * theme.length) % 90 + 10;
            const randomVariance = Math.random() * 20;
            stats[attr] = Math.floor((base + randomVariance) % 99 + 1); 
        });
        resolve(stats);
    });
    return statData;
}

async function fetchCharactersJikan(theme){
  try{
    const q = encodeURIComponent(theme);
    const s = await fetch(`https://api.jikan.moe/v4/anime?q=${q}&limit=1`);
    if (!s.ok) throw new Error('search err');
    const js = await s.json();
    if (!js.data || js.data.length===0) throw new Error('no anime');
    const id = js.data[0].mal_id;
    const c = await fetch(`https://api.jikan.moe/v4/anime/${id}/characters`);
    if (!c.ok) throw new Error('chars err');
    const ch = await c.json();
    if (!ch.data || ch.data.length===0) throw new Error('no chars');
    const arr = ch.data.map(x=>({ name: x.character.name, image: x.character.images?.jpg?.image_url || null }));
    return _.uniqBy(arr,'name');
  }catch(e){ console.warn('jikan fail', e.message); return null; }
}

export async function generateCards(theme, cached){
  const chars = cached || await fetchCharactersJikan(theme);
  const cards = [];
  
  for (let i = 0; i < TOTAL_CARDS; i++){
    let name = `${theme} #${i+1}`;
    let image = `https://picsum.photos/seed/${encodeURIComponent(theme+'|'+i)}/320/420`;

    if (chars && chars.length > 0) {
        const base = chars[i % chars.length];
        name = chars.length >= TOTAL_CARDS ? base.name : `${base.name}${i < chars.length ? '' : ' #' + Math.floor(i/chars.length)}`;
        image = base.image || image;
    }
    
    const stats = await generatePlausibleStats(name, theme);
    
    cards.push({ id: generateId(), name, image, stats });
  }
  return cards;
}