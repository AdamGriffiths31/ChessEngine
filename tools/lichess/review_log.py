"""Review an engine move log (engine.jsonl) against Stockfish.

For every searched move, measures the centipawn loss against Stockfish's best
move. Moves that lost a lot, or whose best move kept changing between
iterations, are printed with Stockfish's top lines and the engine's own
iteration trace, so a margin between candidate moves can be read directly.

    python tools/lichess/review_log.py LOG STOCKFISH_EXE [--depth 16] [--loss 50] [--flips 3]
"""

import argparse
import json
import os
import subprocess

MATE = 100000


class Stockfish:
    def __init__(self, path):
        self.p = subprocess.Popen([os.path.abspath(path)],stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True, bufsize=1)
        self.send("uci")
        self.send("setoption name Threads value 1")
        self.send("isready")
        self.wait("readyok")

    def send(self, line):
        self.p.stdin.write(line + "\n")
        self.p.stdin.flush()

    def wait(self, prefix):
        while True:
            line = self.p.stdout.readline()
            if line.startswith(prefix):
                return line

    def analyse(self, fen, moves, depth, multipv):
        """Return [(move, score)] best first, scores for the side to move."""
        self.send(f"setoption name MultiPV value {multipv}")
        self.send("position fen " + fen + (" moves " + moves if moves else ""))
        self.send(f"go depth {depth}")
        lines = {}
        while True:
            line = self.p.stdout.readline()
            if line.startswith("bestmove"):
                break
            if not line.startswith("info") or " score " not in line or " pv " not in line:
                continue
            t = line.split()
            idx = int(t[t.index("multipv") + 1]) if "multipv" in t else 1
            i = t.index("score")
            val = int(t[i + 2])
            score = val if t[i + 1] == "cp" else (MATE - abs(val)) * (1 if val > 0 else -1)
            lines[idx] = (t[t.index("pv") + 1], score)
        return [lines[k] for k in sorted(lines)]

    def close(self):
        self.send("quit")


def flips(iterations):
    return sum(1 for a, b in zip(iterations, iterations[1:]) if a["best_move"] != b["best_move"])


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("log")
    ap.add_argument("stockfish")
    ap.add_argument("--depth", type=int, default=16)
    ap.add_argument("--loss", type=int, default=50, help="flag moves losing at least this many centipawns")
    ap.add_argument("--flips", type=int, default=3, help="flag searches whose best move changed at least this often")
    args = ap.parse_args()

    sf = Stockfish(args.stockfish)
    moves = []
    for line in open(args.log, encoding="utf-8"):
        r = json.loads(line)
        if r["msg"] == "move" and not r["book"]["used"]:
            moves.append(r)

    total = agree = 0
    flagged = []
    for r in moves:
        top = sf.analyse(r["fen"], "", args.depth, 1)
        if not top or abs(top[0][1]) >= MATE - 1000:
            continue  # mate scores make centipawn loss meaningless
        best_move, best_score = top[0]
        if r["move"] == best_move:
            loss = 0
        else:
            after = sf.analyse(r["fen"], r["move"], args.depth, 1)
            loss = max(0, best_score + after[0][1]) if after else 0
        total += 1
        agree += r["move"] == best_move
        its = r["search"]["iterations"]
        n_flips = flips(its)
        if loss >= args.loss or n_flips >= args.flips:
            flagged.append((r, loss, n_flips))

    print(f"{total} searched moves; Stockfish d{args.depth} agrees with the engine on {agree}")
    print(f"flagged {len(flagged)} (loss >= {args.loss}cp or >= {args.flips} best-move changes)\n")

    for r, loss, n_flips in flagged:
        lines = sf.analyse(r["fen"], "", args.depth, 3)
        its = r["search"]["iterations"]
        if its and "aspiration_fail_lows" in its[0]:
            retries = sum(i["aspiration_fail_lows"] + i["aspiration_fail_highs"] for i in its)
        else:
            retries = "n/a (log predates the counters)"
        print(f"move {r['move_number']}  played {r['move']}  loss {loss}cp  depth {r['depth']}  "
              f"best-move changes {n_flips}  aspiration retries {retries}")
        print("  stockfish: " + "   ".join(f"{m} {s:+d}" for m, s in lines))
        scores = dict(lines)
        if r["move"] in scores:
            print(f"  played move is stockfish #{[m for m, _ in lines].index(r['move']) + 1}, "
                  f"{lines[0][1] - scores[r['move']]}cp behind the best")
        print("  engine by depth: " + " ".join(f"{i['depth']}:{i['best_move']}({i['score_cp']:+d})" for i in its[-8:]))
        print(f"  engine search score {r['score_cp_stm']:+d}, stockfish best {lines[0][1]:+d}\n")

    sf.close()


if __name__ == "__main__":
    main()
