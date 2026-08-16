// lib/db.js - CORRECTED
import { Low } from 'lowdb'; // Changed to import
import { JSONFile } from 'lowdb/node'; // Changed to import
import path from 'path'; // Changed to import
import fs from 'fs'; // Changed to import

const dbFile = path.join(process.cwd(), 'db.json');
if (!fs.existsSync(dbFile)) fs.writeFileSync(dbFile, JSON.stringify({ lobbies: {}, apiCache: {} }, null, 2));
const adapter = new JSONFile(dbFile);
const db = new Low(adapter);
async function init(){ await db.read(); db.data ||= { lobbies: {}, apiCache: {} }; await db.write(); }
async function getDb(){ if (!db.data) await init(); return db; }
export { init, getDb }; // Changed module.exports to export