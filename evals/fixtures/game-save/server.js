// EVAL FIXTURE — vulnerable on purpose. Do not copy.
const express = require('express');
const db = require('./db');
const { requireAuth } = require('./auth');

const app = express();
app.use(express.json());

app.get('/api/save', requireAuth, async (req, res) => {
  const save = await db.saves.findOne({ userId: req.user.id });
  res.json(save);
});

app.put('/api/save', requireAuth, async (req, res) => {
  const current = await db.saves.findOne({ userId: req.user.id });
  if (current && req.body.updatedAt !== current.updatedAt) {
    return res.status(409).json({ error: 'Save is stale, reload and retry' });
  }

  const next = { ...req.body, userId: req.user.id, updatedAt: new Date().toISOString() };
  await db.saves.replaceOne({ userId: req.user.id }, next, { upsert: true });
  res.json(next);
});

app.post('/api/shop/buy', requireAuth, async (req, res) => {
  const { itemId, price } = req.body;
  const save = await db.saves.findOne({ userId: req.user.id });

  if (save.coins < price) {
    return res.status(400).json({ error: 'Not enough coins' });
  }

  save.coins -= price;
  save.inventory.push(itemId);
  await db.saves.replaceOne({ userId: req.user.id }, save);
  res.json({ coins: save.coins, inventory: save.inventory });
});

app.post('/api/leaderboard', requireAuth, async (req, res) => {
  const { playerId, score, level } = req.body;
  await db.scores.insertOne({
    playerId,
    score,
    level,
    achievedAt: req.body.achievedAt || Date.now(),
  });
  res.json({ ok: true });
});

app.post('/api/daily-reward', requireAuth, async (req, res) => {
  const last = await db.rewards.findOne({ userId: req.user.id, day: req.body.day });
  if (last) {
    return res.status(400).json({ error: 'Already claimed today' });
  }

  await db.rewards.insertOne({ userId: req.user.id, day: req.body.day });
  await db.saves.updateOne({ userId: req.user.id }, { $inc: { coins: 100 } });
  res.json({ ok: true });
});

module.exports = app;
