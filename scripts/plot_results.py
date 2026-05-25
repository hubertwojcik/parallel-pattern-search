import csv
import os
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker

ROOT = os.path.join(os.path.dirname(__file__), "..")
RESULTS = os.path.join(ROOT, "benchmarks", "results")
PLOTS   = os.path.join(ROOT, "benchmarks", "plots")
os.makedirs(PLOTS, exist_ok=True)

COLORS = {
    "seq":        "#555555",
    "goroutines": "#2196F3",
    "mpi":        "#4CAF50",
    "pfac":       "#FF5722",
}
LABELS = {
    "seq":        "Sekwencyjny",
    "goroutines": "Goroutines (OpenMP ekw.)",
    "mpi":        "MPI",
    "pfac":       "PFAC/OpenCL (GPU)",
}


def read_csv(name):
    path = os.path.join(RESULTS, name)
    rows = []
    with open(path) as f:
        for r in csv.DictReader(f):
            r["workers"]        = int(r["workers"])
            r["text_mb"]        = int(r["text_mb"])
            r["patterns"]       = int(r["patterns"])
            r["throughput_gbs"] = float(r["throughput_gbs"])
            rows.append(r)
    return rows


def filter_rows(rows, **kw):
    result = rows
    for k, v in kw.items():
        result = [r for r in result if r[k] == v]
    return result


def save(fig, name):
    path = os.path.join(PLOTS, name)
    fig.savefig(path, dpi=150, bbox_inches="tight")
    plt.close(fig)
    print(f"zapisano {path}")


# ── 1. Przyspieszenie (speedup) ────────────────────────────────────────────
workers_data = read_csv("scaling_workers.csv")

fig, ax = plt.subplots(figsize=(7, 5))
ax.set_title("Przyspieszenie względem 1 wątku/procesu\n(10 MB, 100 wzorców)", fontsize=13)

for impl, worker_vals in [("goroutines", [1, 2, 4, 8]), ("mpi", [1, 2, 4])]:
    rows = sorted(filter_rows(workers_data, impl=impl), key=lambda r: r["workers"])
    t1 = rows[0]["throughput_gbs"]
    xs = [r["workers"] for r in rows]
    ys = [r["throughput_gbs"] / t1 for r in rows]
    ax.plot(xs, ys, marker="o", label=LABELS[impl], color=COLORS[impl], linewidth=2)

max_w = 8
ax.plot([1, max_w], [1, max_w], "--", color="gray", linewidth=1, label="Idealne")
ax.set_xlabel("Liczba wątków / procesów")
ax.set_ylabel("Przyspieszenie S(p)")
ax.set_xticks([1, 2, 4, 8])
ax.legend()
ax.grid(True, alpha=0.3)
save(fig, "speedup.png")

# ── 2. Efektywność ─────────────────────────────────────────────────────────
fig, ax = plt.subplots(figsize=(7, 5))
ax.set_title("Efektywność zrównoleglenia\n(10 MB, 100 wzorców)", fontsize=13)

for impl, worker_vals in [("goroutines", [1, 2, 4, 8]), ("mpi", [1, 2, 4])]:
    rows = sorted(filter_rows(workers_data, impl=impl), key=lambda r: r["workers"])
    t1 = rows[0]["throughput_gbs"]
    xs = [r["workers"] for r in rows]
    ys = [(r["throughput_gbs"] / t1) / r["workers"] for r in rows]
    ax.plot(xs, ys, marker="o", label=LABELS[impl], color=COLORS[impl], linewidth=2)

ax.axhline(1.0, linestyle="--", color="gray", linewidth=1, label="Idealna")
ax.set_xlabel("Liczba wątków / procesów")
ax.set_ylabel("Efektywność E(p) = S(p)/p")
ax.set_xticks([1, 2, 4, 8])
ax.set_ylim(0, 1.3)
ax.legend()
ax.grid(True, alpha=0.3)
save(fig, "efficiency.png")

# ── 3. Porównanie najlepszych wyników (bar chart) ──────────────────────────
text_data = read_csv("scaling_text.csv")

fig, ax = plt.subplots(figsize=(8, 5))
ax.set_title("Porównanie najlepszych wyników — przepustowość\n(100 MB, 100 wzorców)", fontsize=13)

configs = [
    ("seq",        1),
    ("goroutines", 4),
    ("mpi",        4),
    ("pfac",       1),
]
labels_bar = []
values_bar = []
colors_bar = []

for impl, w in configs:
    rows = filter_rows(text_data, impl=impl, workers=w, text_mb=100, patterns=100)
    if rows:
        values_bar.append(rows[0]["throughput_gbs"])
        labels_bar.append(LABELS[impl])
        colors_bar.append(COLORS[impl])

bars = ax.bar(labels_bar, values_bar, color=colors_bar, edgecolor="white", width=0.5)
for bar, val in zip(bars, values_bar):
    ax.text(bar.get_x() + bar.get_width() / 2, bar.get_height() + 0.05,
            f"{val:.2f} GB/s", ha="center", va="bottom", fontsize=10)

ax.set_ylabel("Przepustowość (GB/s)")
ax.set_ylim(0, max(values_bar) * 1.25)
ax.grid(True, axis="y", alpha=0.3)
save(fig, "comparison.png")

# ── 4. Skalowanie rozmiaru tekstu ──────────────────────────────────────────
fig, ax = plt.subplots(figsize=(7, 5))
ax.set_title("Przepustowość vs rozmiar tekstu\n(100 wzorców, 4 wątki/procesy)", fontsize=13)

for impl in ["seq", "goroutines", "mpi", "pfac"]:
    w = 1 if impl in ("seq", "pfac") else 4
    rows = sorted(filter_rows(text_data, impl=impl, workers=w, patterns=100),
                  key=lambda r: r["text_mb"])
    if rows:
        xs = [r["text_mb"] for r in rows]
        ys = [r["throughput_gbs"] for r in rows]
        ax.plot(xs, ys, marker="o", label=LABELS[impl], color=COLORS[impl], linewidth=2)

ax.set_xlabel("Rozmiar tekstu (MB)")
ax.set_ylabel("Przepustowość (GB/s)")
ax.set_xticks([10, 100])
ax.legend()
ax.grid(True, alpha=0.3)
save(fig, "scaling_text.png")

# ── 5. Skalowanie liczby wzorców ───────────────────────────────────────────
pat_data = read_csv("scaling_patterns.csv")

fig, ax = plt.subplots(figsize=(7, 5))
ax.set_title("Przepustowość vs liczba wzorców\n(10 MB tekstu, 4 wątki/procesy)", fontsize=13)

for impl in ["seq", "goroutines", "mpi", "pfac"]:
    w = 1 if impl in ("seq", "pfac") else 4
    rows = sorted(filter_rows(pat_data, impl=impl, workers=w, text_mb=10),
                  key=lambda r: r["patterns"])
    if rows:
        xs = [r["patterns"] for r in rows]
        ys = [r["throughput_gbs"] for r in rows]
        ax.plot(xs, ys, marker="o", label=LABELS[impl], color=COLORS[impl], linewidth=2)

ax.set_xlabel("Liczba wzorców")
ax.set_ylabel("Przepustowość (GB/s)")
ax.set_xscale("log")
ax.xaxis.set_major_formatter(ticker.ScalarFormatter())
ax.set_xticks([10, 100, 500, 1000])
ax.legend()
ax.grid(True, alpha=0.3)
save(fig, "scaling_patterns.png")

print("\nWszystkie wykresy zapisane w benchmarks/plots/")
