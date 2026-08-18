-- Dev seed data. Runs automatically after 001_init.sql on a fresh volume
-- (the entrypoint executes files in filename order). Delete this file to
-- start with an empty database.

INSERT INTO users (username, password_hash) VALUES
  ('nairwolf', 'placeholder_hash');

INSERT INTO cards (user_id, card_type, fen, notes, card_details) VALUES
  (
    1,
    'tactic',
    'r1bqkbnr/pppp1Qpp/2n5/4p3/2B1P3/8/PPPP1PPP/RNB1K1NR b KQkq - 0 3',
    'vs. IM_Someone, blitz, 2026-06-02',
    '{"accepted_solution": "Kxf7"}'
  ),
  (
    1,
    'blunder',
    'r1bqk2r/ppp2ppp/2n2n2/3p4/1b1P4/2N1PN2/PP3PPP/R1BQKB1R w KQkq - 0 7',
    'vs. anon, rapid, 2026-06-05',
    '{"intended_move": "Qxe2", "refutation": "Bxc3+", "correct_alternative": "Be2"}'
  ),
  (
    1,
    'strategy',
    'r2q1rk1/1b1nbppp/p3pn2/1p6/3P4/1BN1PN2/PP3PPP/R1BQ1RK1 w - - 0 12',
    'vs. IM_Someone, classical, 2026-06-08',
    '{"question": "Why is trading the light-squared bishop here wrong?", "answer": "It hands Black the bishop pair with an open long diagonal, and my knight has no good square to compensate."}'
  );

-- One live session and one already expired, so auth middleware can be
-- exercised against both. Real ids are crypto/rand tokens; these are
-- deliberately obvious placeholders.
INSERT INTO sessions (id, user_id, created_at, expires_at) VALUES
  ('dev_session_valid',   1, now(),                     now() + interval '7 days'),
  ('dev_session_expired', 1, now() - interval '8 days', now() - interval '1 day');
